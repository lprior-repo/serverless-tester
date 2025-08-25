package helpers

import (
	"fmt"
	"log/slog"
	"strings"
	"sync"
)

// CleanupFunc represents a cleanup function that can return an error
type CleanupFunc func() error

// CleanupItem represents a single cleanup item with name and function
type CleanupItem struct {
	Name    string
	Cleanup CleanupFunc
}

// CleanupManager manages cleanup functions following Terratest patterns
// Executes cleanup functions in LIFO order (last registered, first executed)
type CleanupManager struct {
	mu       sync.Mutex
	cleanups []CleanupItem
}

// NewCleanupManager creates a new cleanup manager
func NewCleanupManager() *CleanupManager {
	return &CleanupManager{
		cleanups: make([]CleanupItem, 0),
	}
}

// Register adds a cleanup function to the manager
func (c *CleanupManager) Register(name string, cleanup CleanupFunc) {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	slog.Info("Registering cleanup function",
		"name", name,
		"totalCleanups", len(c.cleanups)+1,
	)
	
	c.cleanups = append(c.cleanups, CleanupItem{
		Name:    name,
		Cleanup: cleanup,
	})
}

// ExecuteAll executes all registered cleanup functions in LIFO order
// Continues execution even if individual cleanups fail, collecting all errors
func (c *CleanupManager) ExecuteAll() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	if len(c.cleanups) == 0 {
		slog.Info("No cleanup functions to execute")
		return nil
	}
	
	slog.Info("Executing cleanup functions",
		"totalCleanups", len(c.cleanups),
	)
	
	var errors []string
	
	// Execute in reverse order (LIFO)
	for i := len(c.cleanups) - 1; i >= 0; i-- {
		item := c.cleanups[i]
		
		slog.Info("Executing cleanup function",
			"name", item.Name,
			"remainingCleanups", i,
		)
		
		if err := item.Cleanup(); err != nil {
			errorMsg := fmt.Sprintf("cleanup '%s' failed: %v", item.Name, err)
			errors = append(errors, errorMsg)
			
			slog.Error("Cleanup function failed",
				"name", item.Name,
				"error", err.Error(),
			)
		} else {
			slog.Info("Cleanup function succeeded",
				"name", item.Name,
			)
		}
	}
	
	if len(errors) > 0 {
		return fmt.Errorf("cleanup failures: %s", strings.Join(errors, "; "))
	}
	
	slog.Info("All cleanup functions executed successfully")
	return nil
}

// Reset clears all registered cleanup functions
func (c *CleanupManager) Reset() {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	slog.Info("Resetting cleanup manager",
		"previousCleanupCount", len(c.cleanups),
	)
	
	c.cleanups = make([]CleanupItem, 0)
}

// Count returns the number of registered cleanup functions
func (c *CleanupManager) Count() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	return len(c.cleanups)
}

// IsEmpty returns true if no cleanup functions are registered
func (c *CleanupManager) IsEmpty() bool {
	return c.Count() == 0
}

// TestingTCleanup represents the minimal testing interface needed for cleanup registration
type TestingTCleanup interface {
	Cleanup(func())
}

// DeferCleanup registers a cleanup function with the test framework
// Following Terratest patterns for test cleanup registration
func DeferCleanup(t TestingTCleanup, name string, cleanup CleanupFunc) {
	slog.Info("Deferring cleanup with test framework",
		"name", name,
	)
	
	t.Cleanup(func() {
		slog.Info("Executing deferred cleanup",
			"name", name,
		)
		
		if err := cleanup(); err != nil {
			slog.Error("Deferred cleanup failed",
				"name", name,
				"error", err.Error(),
			)
		} else {
			slog.Info("Deferred cleanup succeeded",
				"name", name,
			)
		}
	})
}

// TerraformCleanupWrapper wraps a Terraform destroy function for cleanup
// Following Terratest patterns for Terraform resource cleanup
func TerraformCleanupWrapper(name string, terraformDestroy func()) CleanupFunc {
	return func() error {
		slog.Info("Executing Terraform cleanup",
			"name", name,
		)
		
		defer func() {
			if r := recover(); r != nil {
				slog.Error("Terraform cleanup panicked",
					"name", name,
					"panic", r,
				)
			}
		}()
		
		terraformDestroy()
		
		slog.Info("Terraform cleanup completed",
			"name", name,
		)
		
		return nil
	}
}

// AWSResourceCleanupWrapper wraps an AWS resource deletion function for cleanup
// Following Terratest patterns for AWS resource cleanup
func AWSResourceCleanupWrapper(resourceName, serviceName string, awsCleanup func() error) CleanupFunc {
	return func() error {
		slog.Info("Executing AWS resource cleanup",
			"resourceName", resourceName,
			"serviceName", serviceName,
		)
		
		err := awsCleanup()
		if err != nil {
			slog.Error("AWS resource cleanup failed",
				"resourceName", resourceName,
				"serviceName", serviceName,
				"error", err.Error(),
			)
			return fmt.Errorf("failed to clean up AWS %s resource '%s': %w", serviceName, resourceName, err)
		}
		
		slog.Info("AWS resource cleanup completed",
			"resourceName", resourceName,
			"serviceName", serviceName,
		)
		
		return nil
	}
}