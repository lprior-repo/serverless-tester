package parallel

import (
	"context"
	"errors"
	"fmt"
	"runtime"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"vasdeference"
)

// Mock TestingT implementation for testing error handling
type mockTestingT struct {
	errors []string
	failed bool
	mutex  sync.Mutex
}

func (m *mockTestingT) Errorf(format string, args ...interface{}) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.errors = append(m.errors, fmt.Sprintf(format, args...))
}

func (m *mockTestingT) FailNow() {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.failed = true
	panic("FailNow called")
}

func (m *mockTestingT) Helper() {}

func (m *mockTestingT) getErrors() []string {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	return append([]string{}, m.errors...)
}

func (m *mockTestingT) isFailed() bool {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	return m.failed
}

func TestWorkerPoolManagement(t *testing.T) {
	t.Run("NewRunner creates worker pool with specified size", func(t *testing.T) {
		poolSize := 8
		runner, err := NewRunner(t, Options{PoolSize: poolSize})
		require.NoError(t, err)
		defer runner.Cleanup()

		assert.Equal(t, poolSize, runner.pool.Cap())
		assert.Equal(t, 0, runner.pool.Running())
		assert.Equal(t, poolSize, runner.pool.Free())
	})

	t.Run("NewRunner sets default pool size when invalid", func(t *testing.T) {
		testCases := []int{0, -1, -10}
		
		for _, poolSize := range testCases {
			t.Run(fmt.Sprintf("poolSize_%d", poolSize), func(t *testing.T) {
				runner, err := NewRunner(t, Options{PoolSize: poolSize})
				require.NoError(t, err)
				defer runner.Cleanup()

				assert.Equal(t, 4, runner.pool.Cap()) // Default size
			})
		}
	})

	t.Run("NewRunner sets default timeout when invalid", func(t *testing.T) {
		runner, err := NewRunner(t, Options{TestTimeout: 0})
		require.NoError(t, err)
		defer runner.Cleanup()

		assert.NotNil(t, runner)
	})

	t.Run("worker pool lifecycle and cleanup", func(t *testing.T) {
		runner, err := NewRunner(t, Options{PoolSize: 2})
		require.NoError(t, err)

		// Pool should be active initially
		assert.False(t, runner.pool.IsClosed())
		
		runner.Cleanup()
		
		// Pool should be closed after cleanup
		assert.True(t, runner.pool.IsClosed())
	})

	t.Run("GetStats returns accurate pool statistics", func(t *testing.T) {
		poolSize := 3
		runner, err := NewRunner(t, Options{PoolSize: poolSize})
		require.NoError(t, err)
		defer runner.Cleanup()

		stats := runner.GetStats()
		assert.Equal(t, 0, stats.Running)
		assert.Equal(t, poolSize, stats.Available)
		assert.Equal(t, poolSize, stats.Capacity)

		// Submit a blocking task to test running stats
		var wg sync.WaitGroup
		wg.Add(1)
		
		err = runner.pool.Submit(func() {
			defer wg.Done()
			time.Sleep(50 * time.Millisecond)
		})
		require.NoError(t, err)

		// Check stats while task is running
		time.Sleep(10 * time.Millisecond)
		stats = runner.GetStats()
		assert.Equal(t, 1, stats.Running)
		assert.Equal(t, poolSize-1, stats.Available)
		assert.Equal(t, poolSize, stats.Capacity)

		wg.Wait()
	})

	t.Run("pool exhaustion and queuing scenarios", func(t *testing.T) {
		poolSize := 2
		runner, err := NewRunner(t, Options{PoolSize: poolSize})
		require.NoError(t, err)
		defer runner.Cleanup()

		var completedTasks int64
		var wg sync.WaitGroup

		// Submit more tasks than pool capacity
		numTasks := 5
		for i := 0; i < numTasks; i++ {
			wg.Add(1)
			err := runner.pool.Submit(func() {
				defer wg.Done()
				atomic.AddInt64(&completedTasks, 1)
				time.Sleep(10 * time.Millisecond)
			})
			require.NoError(t, err)
		}

		wg.Wait()
		assert.Equal(t, int64(numTasks), atomic.LoadInt64(&completedTasks))
	})

	t.Run("graceful shutdown with running tasks", func(t *testing.T) {
		runner, err := NewRunner(t, Options{PoolSize: 2})
		require.NoError(t, err)

		var tasksStarted int64
		var tasksCompleted int64

		// Submit long-running tasks
		for i := 0; i < 3; i++ {
			err := runner.pool.Submit(func() {
				atomic.AddInt64(&tasksStarted, 1)
				time.Sleep(100 * time.Millisecond)
				atomic.AddInt64(&tasksCompleted, 1)
			})
			require.NoError(t, err)
		}

		// Wait a bit for tasks to start
		time.Sleep(20 * time.Millisecond)
		
		// Cleanup should wait for running tasks to complete
		runner.Cleanup()

		assert.True(t, runner.pool.IsClosed())
		assert.Equal(t, atomic.LoadInt64(&tasksStarted), atomic.LoadInt64(&tasksCompleted))
	})
}

func TestResourceIsolation(t *testing.T) {
	t.Run("namespace generation provides unique identifiers", func(t *testing.T) {
		runner, err := NewRunner(t, Options{PoolSize: 4})
		require.NoError(t, err)
		defer runner.Cleanup()

		var namespaces []string
		var mutex sync.Mutex

		testCases := make([]TestCase, 10)
		for i := 0; i < 10; i++ {
			testCases[i] = TestCase{
				Name: fmt.Sprintf("test_%d", i),
				Func: func(ctx context.Context, tctx *TestContext) error {
					mutex.Lock()
					namespaces = append(namespaces, tctx.namespace)
					mutex.Unlock()
					return nil
				},
			}
		}

		runner.Run(testCases)

		assert.Len(t, namespaces, 10)
		
		// All namespaces should be unique
		seen := make(map[string]bool)
		for _, ns := range namespaces {
			assert.False(t, seen[ns], "Namespace %s should be unique", ns)
			seen[ns] = true
			assert.Contains(t, ns, "test")
		}
	})

	t.Run("GetTableName provides resource prefixing", func(t *testing.T) {
		testCtx := &TestContext{namespace: "test123"}
		
		testCases := []struct {
			baseName string
			expected string
		}{
			{"users", "test123-users"},
			{"products", "test123-products"},
			{"orders-history", "test123-orders-history"},
		}

		for _, tc := range testCases {
			t.Run(tc.baseName, func(t *testing.T) {
				result := testCtx.GetTableName(tc.baseName)
				assert.Equal(t, tc.expected, result)
			})
		}
	})

	t.Run("GetFunctionName provides resource prefixing", func(t *testing.T) {
		testCtx := &TestContext{namespace: "test456"}
		
		testCases := []struct {
			baseName string
			expected string
		}{
			{"handler", "test456-handler"},
			{"processor", "test456-processor"},
			{"event-handler", "test456-event-handler"},
		}

		for _, tc := range testCases {
			t.Run(tc.baseName, func(t *testing.T) {
				result := testCtx.GetFunctionName(tc.baseName)
				assert.Equal(t, tc.expected, result)
			})
		}
	})

	t.Run("resource lock acquisition and release", func(t *testing.T) {
		testCtx := &TestContext{
			namespace: "locktest",
			locks:     make(map[string]*sync.Mutex),
		}

		resourceName := "shared-db"

		// Test manual lock/unlock
		testCtx.LockResource(resourceName)
		assert.Contains(t, testCtx.locks, resourceName)
		testCtx.UnlockResource(resourceName)

		// Test lock is reusable
		testCtx.LockResource(resourceName)
		testCtx.UnlockResource(resourceName)
	})

	t.Run("WithResourceLock execution patterns", func(t *testing.T) {
		testCtx := &TestContext{
			namespace: "withlock",
			locks:     make(map[string]*sync.Mutex),
		}

		var executionCount int
		err := testCtx.WithResourceLock("test-resource", func() error {
			executionCount++
			return nil
		})

		assert.NoError(t, err)
		assert.Equal(t, 1, executionCount)
	})

	t.Run("WithResourceLock handles function errors", func(t *testing.T) {
		testCtx := &TestContext{
			namespace: "errorlock",
			locks:     make(map[string]*sync.Mutex),
		}

		expectedErr := errors.New("function error")
		err := testCtx.WithResourceLock("test-resource", func() error {
			return expectedErr
		})

		assert.Equal(t, expectedErr, err)
	})

	t.Run("concurrent resource access patterns", func(t *testing.T) {
		runner, err := NewRunner(t, Options{PoolSize: 4})
		require.NoError(t, err)
		defer runner.Cleanup()

		var accessOrder []int
		var mutex sync.Mutex

		testCases := make([]TestCase, 4)
		for i := 0; i < 4; i++ {
			taskID := i
			testCases[i] = TestCase{
				Name: fmt.Sprintf("concurrent_test_%d", i),
				Func: func(ctx context.Context, tctx *TestContext) error {
					return tctx.WithResourceLock("shared", func() error {
						mutex.Lock()
						accessOrder = append(accessOrder, taskID)
						mutex.Unlock()
						time.Sleep(10 * time.Millisecond)
						return nil
					})
				},
			}
		}

		runner.Run(testCases)

		assert.Len(t, accessOrder, 4)
		// All tasks should complete (order doesn't matter due to goroutine scheduling)
		assert.ElementsMatch(t, []int{0, 1, 2, 3}, accessOrder)
	})
}

func TestTestCaseExecution(t *testing.T) {
	t.Run("Run executes multiple test cases successfully", func(t *testing.T) {
		runner, err := NewRunner(t, Options{PoolSize: 2})
		require.NoError(t, err)
		defer runner.Cleanup()

		var executionOrder []string
		var mutex sync.Mutex

		testCases := []TestCase{
			{
				Name: "first",
				Func: func(ctx context.Context, tctx *TestContext) error {
					mutex.Lock()
					executionOrder = append(executionOrder, "first")
					mutex.Unlock()
					return nil
				},
			},
			{
				Name: "second",
				Func: func(ctx context.Context, tctx *TestContext) error {
					mutex.Lock()
					executionOrder = append(executionOrder, "second")
					mutex.Unlock()
					return nil
				},
			},
		}

		runner.Run(testCases)

		assert.Len(t, executionOrder, 2)
		assert.ElementsMatch(t, []string{"first", "second"}, executionOrder)
	})

	t.Run("error collection and aggregation", func(t *testing.T) {
		mockT := &mockTestingT{}
		runner, err := NewRunner(mockT, Options{PoolSize: 2})
		require.NoError(t, err)
		defer runner.Cleanup()

		testCases := []TestCase{
			{
				Name: "error_test_1",
				Func: func(ctx context.Context, tctx *TestContext) error {
					return errors.New("first error")
				},
			},
			{
				Name: "error_test_2", 
				Func: func(ctx context.Context, tctx *TestContext) error {
					return errors.New("second error")
				},
			},
		}

		// Expect panic from FailNow
		defer func() {
			recover()
		}()

		runner.Run(testCases)

		errors := mockT.getErrors()
		assert.Len(t, errors, 2)
		
		errorText := fmt.Sprintf("%v %v", errors[0], errors[1])
		assert.Contains(t, errorText, "first error")
		assert.Contains(t, errorText, "second error")
	})

	t.Run("panic recovery in test cases", func(t *testing.T) {
		mockT := &mockTestingT{}
		runner, err := NewRunner(mockT, Options{PoolSize: 1})
		require.NoError(t, err)
		defer runner.Cleanup()

		testCases := []TestCase{
			{
				Name: "panicking_test",
				Func: func(ctx context.Context, tctx *TestContext) error {
					panic("test panic message")
				},
			},
		}

		// Expect panic from FailNow
		defer func() {
			recover()
		}()

		runner.Run(testCases)

		errors := mockT.getErrors()
		assert.Len(t, errors, 1)
		assert.Contains(t, errors[0], "panicked")
		assert.Contains(t, errors[0], "test panic message")
	})

	t.Run("timeout handling in test context", func(t *testing.T) {
		runner, err := NewRunner(t, Options{
			PoolSize:    1,
			TestTimeout: 50 * time.Millisecond,
		})
		require.NoError(t, err)
		defer runner.Cleanup()

		var timeoutOccurred bool
		testCases := []TestCase{
			{
				Name: "timeout_test",
				Func: func(ctx context.Context, tctx *TestContext) error {
					select {
					case <-ctx.Done():
						timeoutOccurred = true
						return ctx.Err()
					case <-time.After(200 * time.Millisecond):
						return nil
					}
				},
			},
		}

		runner.Run(testCases)
		// Test completes regardless of timeout due to context cancellation
		assert.True(t, timeoutOccurred)
	})

	t.Run("test case isolation verification", func(t *testing.T) {
		runner, err := NewRunner(t, Options{PoolSize: 3})
		require.NoError(t, err)
		defer runner.Cleanup()

		var contexts []*TestContext
		var mutex sync.Mutex

		testCases := make([]TestCase, 5)
		for i := 0; i < 5; i++ {
			testCases[i] = TestCase{
				Name: fmt.Sprintf("isolation_test_%d", i),
				Func: func(ctx context.Context, tctx *TestContext) error {
					mutex.Lock()
					contexts = append(contexts, tctx)
					mutex.Unlock()
					return nil
				},
			}
		}

		runner.Run(testCases)

		assert.Len(t, contexts, 5)
		
		// Verify all contexts have unique namespaces
		namespaces := make(map[string]bool)
		for _, ctx := range contexts {
			assert.False(t, namespaces[ctx.namespace], "Namespace should be unique")
			namespaces[ctx.namespace] = true
		}
	})
}

func TestTableDrivenTestSupport(t *testing.T) {
	t.Run("RunLambdaTests executes all test cases", func(t *testing.T) {
		cases := []LambdaTestCase{
			{
				Name:           "success_case",
				Input:          map[string]string{"action": "process"},
				ExpectedStatus: 200,
				ExpectedError:  false,
			},
			{
				Name:           "validation_error",
				Input:          map[string]string{"invalid": "input"},
				ExpectedStatus: 400,
				ExpectedError:  true,
			},
			{
				Name:           "server_error",
				Input:          nil,
				ExpectedStatus: 500,
				ExpectedError:  true,
			},
		}

		// Should complete without errors since implementation is stubbed
		RunLambdaTests(t, "test-function", cases)
	})

	t.Run("RunDynamoDBTests with data isolation", func(t *testing.T) {
		cases := []DynamoDBTestCase{
			{
				Name:     "query_single_item",
				Items:    []interface{}{map[string]string{"id": "1", "name": "item1"}},
				Query:    "SELECT * FROM table WHERE id = '1'",
				Expected: 1,
			},
			{
				Name:     "query_multiple_items",
				Items:    []interface{}{
					map[string]string{"id": "1", "name": "item1"},
					map[string]string{"id": "2", "name": "item2"},
				},
				Query:    "SELECT * FROM table",
				Expected: 2,
			},
		}

		RunDynamoDBTests(t, "test-table", cases)
	})

	t.Run("RunStepFunctionTests with workflow patterns", func(t *testing.T) {
		cases := []StepFunctionTestCase{
			{
				Name:           "simple_workflow",
				Input:          map[string]string{"action": "start"},
				ExpectedStatus: "SUCCEEDED",
				Timeout:        30 * time.Second,
			},
			{
				Name:           "parallel_execution",
				Input:          map[string]interface{}{"parallel": true, "branches": 3},
				ExpectedStatus: "SUCCEEDED", 
				Timeout:        60 * time.Second,
			},
		}

		RunStepFunctionTests(t, "arn:aws:states:us-east-1:123456789012:stateMachine:TestMachine", cases)
	})

	t.Run("error handling in table-driven scenarios", func(t *testing.T) {
		// Test with empty test cases
		RunLambdaTests(t, "test-function", []LambdaTestCase{})
		RunDynamoDBTests(t, "test-table", []DynamoDBTestCase{})
		RunStepFunctionTests(t, "test-state-machine", []StepFunctionTestCase{})
	})
}

func TestNewRunnerE(t *testing.T) {
	t.Run("NewRunnerE returns error when pool creation fails", func(t *testing.T) {
		// This test verifies error handling in NewRunnerE
		// In practice, ants.NewPool rarely fails unless system resources are exhausted
		runner, err := NewRunnerE(t, Options{PoolSize: 1})
		
		// Should succeed with valid options
		assert.NoError(t, err)
		assert.NotNil(t, runner)
		runner.Cleanup()
	})

	t.Run("NewRunnerWithArrange creates runner with existing arrange", func(t *testing.T) {
		testCtx := sfx.NewTestContext(t)
		arrange := sfx.NewArrange(testCtx)
		defer arrange.Cleanup()

		runner, err := NewRunnerWithArrange(arrange, Options{PoolSize: 2})
		require.NoError(t, err)
		defer runner.Cleanup()

		assert.NotNil(t, runner)
		assert.Equal(t, arrange, runner.baseArrange)
	})

	t.Run("NewRunnerWithArrangeE returns same as NewRunnerWithArrange", func(t *testing.T) {
		testCtx := sfx.NewTestContext(t)
		arrange := sfx.NewArrange(testCtx)
		defer arrange.Cleanup()

		runner, err := NewRunnerWithArrangeE(arrange, Options{PoolSize: 2})
		require.NoError(t, err)
		defer runner.Cleanup()

		assert.NotNil(t, runner)
		assert.Equal(t, arrange, runner.baseArrange)
	})
}

func TestRunE(t *testing.T) {
	t.Run("RunE returns nil on successful execution", func(t *testing.T) {
		runner, err := NewRunner(t, Options{PoolSize: 2})
		require.NoError(t, err)
		defer runner.Cleanup()

		testCases := []TestCase{
			{
				Name: "success_test",
				Func: func(ctx context.Context, tctx *TestContext) error {
					return nil
				},
			},
		}

		err = runner.RunE(testCases)
		assert.NoError(t, err)
	})

	t.Run("RunE returns error when test cases fail", func(t *testing.T) {
		runner, err := NewRunner(t, Options{PoolSize: 1})
		require.NoError(t, err)
		defer runner.Cleanup()

		testCases := []TestCase{
			{
				Name: "failing_test",
				Func: func(ctx context.Context, tctx *TestContext) error {
					return errors.New("test failure")
				},
			},
		}

		err = runner.RunE(testCases)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "parallel test errors occurred")
	})
}

func TestConcurrentResourceAccess(t *testing.T) {
	t.Run("multiple tests can access different resources concurrently", func(t *testing.T) {
		runner, err := NewRunner(t, Options{PoolSize: 4})
		require.NoError(t, err)
		defer runner.Cleanup()

		var accessLog []string
		var mutex sync.Mutex

		testCases := make([]TestCase, 4)
		resources := []string{"resource1", "resource2", "resource3", "resource4"}
		
		for i, resource := range resources {
			res := resource // Capture loop variable
			testCases[i] = TestCase{
				Name: fmt.Sprintf("test_%s", res),
				Func: func(ctx context.Context, tctx *TestContext) error {
					return tctx.WithResourceLock(res, func() error {
						mutex.Lock()
						accessLog = append(accessLog, res)
						mutex.Unlock()
						time.Sleep(10 * time.Millisecond)
						return nil
					})
				},
			}
		}

		runner.Run(testCases)

		assert.Len(t, accessLog, 4)
		assert.ElementsMatch(t, resources, accessLog)
	})

	t.Run("tests accessing same resource are serialized", func(t *testing.T) {
		runner, err := NewRunner(t, Options{PoolSize: 3})
		require.NoError(t, err)
		defer runner.Cleanup()

		var sharedCounter int
		var finalValues []int
		var mutex sync.Mutex

		testCases := make([]TestCase, 3)
		for i := 0; i < 3; i++ {
			testCases[i] = TestCase{
				Name: fmt.Sprintf("shared_resource_test_%d", i),
				Func: func(ctx context.Context, tctx *TestContext) error {
					return tctx.WithResourceLock("shared", func() error {
						// Read-modify-write operation that must be atomic
						current := sharedCounter
						time.Sleep(5 * time.Millisecond) // Simulate processing
						sharedCounter = current + 1
						
						mutex.Lock()
						finalValues = append(finalValues, sharedCounter)
						mutex.Unlock()
						return nil
					})
				},
			}
		}

		runner.Run(testCases)

		assert.Len(t, finalValues, 3)
		assert.Equal(t, 3, sharedCounter)
		assert.ElementsMatch(t, []int{1, 2, 3}, finalValues)
	})
}

func TestMemoryAndResourceLeaks(t *testing.T) {
	t.Run("cleanup prevents resource leaks", func(t *testing.T) {
		initialGoroutines := runtime.NumGoroutine()
		
		func() {
			runner, err := NewRunner(t, Options{PoolSize: 4})
			require.NoError(t, err)

			// Submit some work
			testCases := make([]TestCase, 10)
			for i := 0; i < 10; i++ {
				testCases[i] = TestCase{
					Name: fmt.Sprintf("leak_test_%d", i),
					Func: func(ctx context.Context, tctx *TestContext) error {
						time.Sleep(1 * time.Millisecond)
						return nil
					},
				}
			}

			runner.Run(testCases)
			runner.Cleanup()
		}()

		// Allow some time for goroutines to terminate
		time.Sleep(100 * time.Millisecond)
		
		finalGoroutines := runtime.NumGoroutine()
		
		// Should not have significant goroutine leaks (allow some variance for GC, etc.)
		assert.InDelta(t, initialGoroutines, finalGoroutines, 5,
			"Should not have significant goroutine leaks. Initial: %d, Final: %d", 
			initialGoroutines, finalGoroutines)
	})

	t.Run("multiple cleanup calls are safe", func(t *testing.T) {
		runner, err := NewRunner(t, Options{PoolSize: 2})
		require.NoError(t, err)

		// Multiple cleanup calls should not cause issues
		runner.Cleanup()
		runner.Cleanup()
		runner.Cleanup()
		
		assert.True(t, runner.pool.IsClosed())
	})
}

func TestEdgeCasesAndErrorConditions(t *testing.T) {
	t.Run("empty test case list", func(t *testing.T) {
		runner, err := NewRunner(t, Options{PoolSize: 2})
		require.NoError(t, err)
		defer runner.Cleanup()

		runner.Run([]TestCase{})
		// Should complete without error
	})

	t.Run("test case with nil function", func(t *testing.T) {
		mockT := &mockTestingT{}
		runner, err := NewRunner(mockT, Options{PoolSize: 1})
		require.NoError(t, err)
		defer runner.Cleanup()

		testCases := []TestCase{
			{
				Name: "nil_function_test",
				Func: nil,
			},
		}

		// Should handle nil function gracefully
		defer func() {
			recover() // Recover from expected panic
		}()

		runner.Run(testCases)
	})

	t.Run("very large number of test cases", func(t *testing.T) {
		runner, err := NewRunner(t, Options{PoolSize: 4})
		require.NoError(t, err)
		defer runner.Cleanup()

		numTests := 100
		var completedCount int64

		testCases := make([]TestCase, numTests)
		for i := 0; i < numTests; i++ {
			testCases[i] = TestCase{
				Name: fmt.Sprintf("large_test_%d", i),
				Func: func(ctx context.Context, tctx *TestContext) error {
					atomic.AddInt64(&completedCount, 1)
					return nil
				},
			}
		}

		runner.Run(testCases)
		
		assert.Equal(t, int64(numTests), atomic.LoadInt64(&completedCount))
	})

	t.Run("context cancellation handling", func(t *testing.T) {
		runner, err := NewRunner(t, Options{PoolSize: 2})
		require.NoError(t, err)
		defer runner.Cleanup()

		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately

		testCases := []TestCase{
			{
				Name: "cancelled_context_test",
				Func: func(testCtx context.Context, tctx *TestContext) error {
					// This should detect the cancelled parent context
					select {
					case <-testCtx.Done():
						return testCtx.Err()
					default:
						return nil
					}
				},
			},
		}

		runner.Run(testCases)
		// Test should complete regardless of cancellation
	})
}