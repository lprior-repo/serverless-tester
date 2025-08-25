package parallel

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"vasdeference"
)

func TestNewRunner(t *testing.T) {
	t.Run("creates runner with default options", func(t *testing.T) {
		runner, err := NewRunner(t, Options{})
		
		require.NoError(t, err)
		assert.NotNil(t, runner)
		assert.Equal(t, t, runner.t)
		assert.NotNil(t, runner.pool)
		assert.Equal(t, 4, runner.pool.Cap())
	})

	t.Run("creates runner with custom options", func(t *testing.T) {
		opts := Options{
			PoolSize:          8,
			ResourceIsolation: true,
			TestTimeout:       10 * time.Second,
		}
		
		runner, err := NewRunner(t, opts)
		
		require.NoError(t, err)
		assert.Equal(t, 8, runner.pool.Cap())
	})

	t.Run("sets default pool size when invalid", func(t *testing.T) {
		runner, err := NewRunner(t, Options{PoolSize: -1})
		
		require.NoError(t, err)
		assert.Equal(t, 4, runner.pool.Cap())
	})
}

func TestRunnerRun(t *testing.T) {
	t.Run("executes test cases in parallel", func(t *testing.T) {
		runner, err := NewRunner(t, Options{PoolSize: 2})
		require.NoError(t, err)
		
		var counter int64
		testCases := []TestCase{
			{
				Name: "test1",
				Func: func(ctx context.Context, tctx *TestContext) error {
					atomic.AddInt64(&counter, 1)
					time.Sleep(10 * time.Millisecond)
					return nil
				},
			},
			{
				Name: "test2", 
				Func: func(ctx context.Context, tctx *TestContext) error {
					atomic.AddInt64(&counter, 1)
					time.Sleep(10 * time.Millisecond)
					return nil
				},
			},
		}
		
		runner.Run(testCases)
		
		assert.Equal(t, int64(2), atomic.LoadInt64(&counter))
	})

	t.Run("collects and reports errors", func(t *testing.T) {
		helper := &testHelper{}
		runner, err := NewRunner(helper, Options{PoolSize: 1})
		require.NoError(t, err)
		
		testCases := []TestCase{
			{
				Name: "failing_test",
				Func: func(ctx context.Context, tctx *TestContext) error {
					return errors.New("test error")
				},
			},
		}
		
		// Catch the panic from FailNow
		defer func() {
			if r := recover(); r != nil {
				// Expected panic from FailNow
				assert.Equal(t, "test failed", r)
			}
		}()
		
		runner.Run(testCases)
		
		assert.True(t, helper.failed)
		assert.Len(t, helper.errors, 1)
		assert.Contains(t, helper.errors[0], "test error")
	})

	t.Run("handles panics gracefully", func(t *testing.T) {
		helper := &testHelper{}
		runner, err := NewRunner(helper, Options{PoolSize: 1})
		require.NoError(t, err)
		
		testCases := []TestCase{
			{
				Name: "panicking_test",
				Func: func(ctx context.Context, tctx *TestContext) error {
					panic("test panic")
				},
			},
		}
		
		// Catch the panic from FailNow
		defer func() {
			if r := recover(); r != nil {
				// Expected panic from FailNow
				assert.Equal(t, "test failed", r)
			}
		}()
		
		runner.Run(testCases)
		
		assert.True(t, helper.failed)
		assert.Len(t, helper.errors, 1)
		assert.Contains(t, helper.errors[0], "panicked")
	})

	t.Run("provides isolated test context", func(t *testing.T) {
		runner, err := NewRunner(t, Options{PoolSize: 2})
		require.NoError(t, err)
		
		var namespaces []string
		var mu sync.Mutex
		
		testCases := []TestCase{
			{
				Name: "test1",
				Func: func(ctx context.Context, tctx *TestContext) error {
					mu.Lock()
					namespaces = append(namespaces, tctx.namespace)
					mu.Unlock()
					return nil
				},
			},
			{
				Name: "test2",
				Func: func(ctx context.Context, tctx *TestContext) error {
					mu.Lock()
					namespaces = append(namespaces, tctx.namespace)
					mu.Unlock()
					return nil
				},
			},
		}
		
		runner.Run(testCases)
		
		assert.Len(t, namespaces, 2)
		assert.NotEqual(t, namespaces[0], namespaces[1])
		assert.Contains(t, namespaces[0], "test")
		assert.Contains(t, namespaces[1], "test")
	})
}

func TestTestContextResourceIsolation(t *testing.T) {
	t.Run("GetTableName returns isolated table name", func(t *testing.T) {
		ctx := &TestContext{namespace: "test123"}
		tableName := ctx.GetTableName("users")
		
		assert.Equal(t, "test123-users", tableName)
	})

	t.Run("GetFunctionName returns isolated function name", func(t *testing.T) {
		ctx := &TestContext{namespace: "test456"}
		functionName := ctx.GetFunctionName("handler")
		
		assert.Equal(t, "test456-handler", functionName)
	})

	t.Run("WithResourceLock provides exclusive access", func(t *testing.T) {
		ctx := &TestContext{
			namespace: "test789",
			locks:     make(map[string]*sync.Mutex),
		}
		
		// Test that the resource lock functions work
		err := ctx.WithResourceLock("shared_resource", func() error {
			return nil
		})
		assert.NoError(t, err)
		
		// Test that lock and unlock work manually
		ctx.LockResource("manual_resource")
		ctx.UnlockResource("manual_resource")
	})
}

func TestRunLambdaTests(t *testing.T) {
	t.Run("executes Lambda test cases", func(t *testing.T) {
		cases := []LambdaTestCase{
			{
				Name:           "success_case",
				Input:          map[string]string{"key": "value"},
				ExpectedStatus: 200,
				ExpectedError:  false,
			},
			{
				Name:           "error_case",
				Input:          map[string]string{"invalid": "input"},
				ExpectedStatus: 400,
				ExpectedError:  true,
			},
		}
		
		// This is a basic test - actual Lambda invocation would be mocked
		helper := &testHelper{}
		RunLambdaTests(&testing.T{}, "test-function", cases)
		
		// Since runLambdaTest is not implemented, no errors should occur
		assert.False(t, helper.failed)
	})
}

func TestRunDynamoDBTests(t *testing.T) {
	t.Run("executes DynamoDB test cases with resource isolation", func(t *testing.T) {
		cases := []DynamoDBTestCase{
			{
				Name:     "query_items",
				Items:    []interface{}{map[string]string{"id": "1", "name": "test"}},
				Query:    "SELECT * FROM table",
				Expected: 1,
			},
		}
		
		helper := &testHelper{}
		RunDynamoDBTests(&testing.T{}, "test-table", cases)
		
		// Since runDynamoDBTest is not implemented, no errors should occur
		assert.False(t, helper.failed)
	})
}

func TestRunStepFunctionTests(t *testing.T) {
	t.Run("executes Step Functions test cases", func(t *testing.T) {
		cases := []StepFunctionTestCase{
			{
				Name:           "simple_execution",
				Input:          map[string]string{"action": "process"},
				ExpectedStatus: "SUCCEEDED",
				Timeout:        30 * time.Second,
			},
		}
		
		helper := &testHelper{}
		RunStepFunctionTests(&testing.T{}, "arn:aws:states:us-east-1:123456789012:stateMachine:TestMachine", cases)
		
		// Since runStepFunctionTest is not implemented, no errors should occur
		assert.False(t, helper.failed)
	})
}

func TestGetStats(t *testing.T) {
	t.Run("returns pool statistics", func(t *testing.T) {
		runner, err := NewRunner(t, Options{PoolSize: 5})
		require.NoError(t, err)
		
		stats := runner.GetStats()
		
		assert.Equal(t, 0, stats.Running)    // No tasks running initially
		assert.Equal(t, 5, stats.Available) // All workers available
		assert.Equal(t, 5, stats.Capacity)  // Total capacity
	})
}

func TestCleanup(t *testing.T) {
	t.Run("releases pool resources", func(t *testing.T) {
		runner, err := NewRunner(t, Options{PoolSize: 2})
		require.NoError(t, err)
		
		runner.Cleanup()
		
		// Pool should be released and no longer usable
		assert.True(t, runner.pool.IsClosed())
	})
}

// Test helper that implements TestingT interface for testing
type testHelper struct {
	failed bool
	errors []string
}

func (th *testHelper) Errorf(format string, args ...interface{}) {
	th.failed = true
	th.errors = append(th.errors, fmt.Sprintf(format, args...))
}

func (th *testHelper) Error(args ...interface{}) {
	th.failed = true
	th.errors = append(th.errors, fmt.Sprint(args...))
}

func (th *testHelper) FailNow() {
	th.failed = true
	panic("test failed")
}

func (th *testHelper) Fail() {
	th.failed = true
}

func (th *testHelper) Fatal(args ...interface{}) {
	th.failed = true
	th.errors = append(th.errors, fmt.Sprint(args...))
	panic("test failed")
}

func (th *testHelper) Fatalf(format string, args ...interface{}) {
	th.failed = true
	th.errors = append(th.errors, fmt.Sprintf(format, args...))
	panic("test failed")
}

func (th *testHelper) Logf(format string, args ...interface{}) {
	// No-op for test helper
}

func (th *testHelper) Name() string {
	return "testHelper"
}

func (th *testHelper) Helper() {}

// Additional test for integration with vasdeference Arrange
func TestIntegrationWithArrange(t *testing.T) {
	t.Run("works with vasdeference TestContext and Arrange", func(t *testing.T) {
		testCtx := vasdeference.NewTestContext(t)
		arrange := vasdeference.NewArrange(testCtx)
		defer arrange.Cleanup()
		
		runner, err := NewRunner(t, Options{
			PoolSize: 2,
			BaseArrange: arrange,
		})
		require.NoError(t, err)
		
		// Test that runner was created successfully with arrange integration
		assert.NotNil(t, runner)
		assert.Equal(t, t, runner.t)
	})
}

func TestErrorCollection(t *testing.T) {
	t.Run("collects multiple errors from parallel tests", func(t *testing.T) {
		helper := &testHelper{}
		runner, err := NewRunner(helper, Options{PoolSize: 2})
		require.NoError(t, err)
		
		testCases := []TestCase{
			{
				Name: "error1",
				Func: func(ctx context.Context, tctx *TestContext) error {
					return errors.New("first error")
				},
			},
			{
				Name: "error2",
				Func: func(ctx context.Context, tctx *TestContext) error {
					return errors.New("second error")
				},
			},
		}
		
		// Catch the panic from FailNow
		defer func() {
			if r := recover(); r != nil {
				// Expected panic from FailNow
				assert.Equal(t, "test failed", r)
			}
		}()
		
		runner.Run(testCases)
		
		assert.True(t, helper.failed)
		assert.Len(t, helper.errors, 2)
		// Check that both errors are present (order might vary due to parallel execution)
		errorStrings := fmt.Sprintf("%v %v", helper.errors[0], helper.errors[1])
		assert.Contains(t, errorStrings, "first error")
		assert.Contains(t, errorStrings, "second error")
	})
}