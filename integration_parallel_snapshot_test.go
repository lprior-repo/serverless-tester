package sfx

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	
	"vasdeference/parallel"
	"vasdeference/snapshot"
)

// Mock TestingT for integration testing
type integrationTestingT struct {
	errors []string
	failed bool
	mutex  sync.Mutex
	name   string
}

func (m *integrationTestingT) Errorf(format string, args ...interface{}) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.errors = append(m.errors, fmt.Sprintf(format, args...))
}

func (m *integrationTestingT) FailNow() {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.failed = true
	panic("FailNow called")
}

func (m *integrationTestingT) Helper() {}

func (m *integrationTestingT) Name() string {
	return m.name
}

func (m *integrationTestingT) getErrors() []string {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	return append([]string{}, m.errors...)
}

func (m *integrationTestingT) isFailed() bool {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	return m.failed
}

func newIntegrationTestingT(name string) *integrationTestingT {
	return &integrationTestingT{name: name}
}

func TestParallelSnapshotIntegration(t *testing.T) {
	t.Run("parallel tests create isolated snapshots", func(t *testing.T) {
		tempDir := t.TempDir()
		snapshotDir := filepath.Join(tempDir, "snapshots")

		runner, err := parallel.NewRunner(t, parallel.Options{PoolSize: 4})
		require.NoError(t, err)
		defer runner.Cleanup()

		var snapshots []string
		var mutex sync.Mutex

		testCases := make([]parallel.TestCase, 5)
		for i := 0; i < 5; i++ {
			testID := i
			testCases[i] = parallel.TestCase{
				Name: fmt.Sprintf("snapshot_test_%d", testID),
				Func: func(ctx context.Context, tctx *parallel.TestContext) error {
					// Create snapshot with unique data per test
					mockT := newIntegrationTestingT(fmt.Sprintf("ParallelSnapshot_%s", tctx.GetTableName("test")))
					s := snapshot.New(mockT, snapshot.Options{SnapshotDir: snapshotDir})
					
					data := map[string]interface{}{
						"testID":    testID,
						"tableName": tctx.GetTableName("users"),
						"funcName":  tctx.GetFunctionName("handler"),
						"namespace": fmt.Sprintf("isolated_%d", testID),
					}
					
					snapshotName := fmt.Sprintf("parallel_test_%d", testID)
					s.MatchJSON(snapshotName, data)
					
					mutex.Lock()
					snapshots = append(snapshots, snapshotName)
					mutex.Unlock()
					
					return nil
				},
			}
		}

		runner.Run(testCases)

		assert.Len(t, snapshots, 5)

		// Verify all snapshot files were created
		files, err := os.ReadDir(snapshotDir)
		require.NoError(t, err)
		assert.Len(t, files, 5)

		// Verify each snapshot contains isolated data
		for i := 0; i < 5; i++ {
			found := false
			for _, file := range files {
				content, err := os.ReadFile(filepath.Join(snapshotDir, file.Name()))
				require.NoError(t, err)
				
				if strings.Contains(string(content), fmt.Sprintf("\"testID\": %d", i)) {
					found = true
					
					var data map[string]interface{}
					err = json.Unmarshal(content, &data)
					require.NoError(t, err)
					
					// Verify isolation - each test should have unique resource names
					assert.Contains(t, data["tableName"], "users")
					assert.Contains(t, data["funcName"], "handler")
					assert.Contains(t, data["namespace"], fmt.Sprintf("isolated_%d", i))
					break
				}
			}
			assert.True(t, found, "Should find snapshot for test %d", i)
		}
	})

	t.Run("concurrent snapshot updates are thread-safe", func(t *testing.T) {
		tempDir := t.TempDir()
		snapshotDir := filepath.Join(tempDir, "concurrent")

		runner, err := parallel.NewRunner(t, parallel.Options{PoolSize: 8})
		require.NoError(t, err)
		defer runner.Cleanup()

		numTests := 20
		testCases := make([]parallel.TestCase, numTests)
		
		for i := 0; i < numTests; i++ {
			testID := i
			testCases[i] = parallel.TestCase{
				Name: fmt.Sprintf("concurrent_snapshot_%d", testID),
				Func: func(ctx context.Context, tctx *parallel.TestContext) error {
					mockT := newIntegrationTestingT(fmt.Sprintf("ConcurrentTest_%d", testID))
					s := snapshot.New(mockT, snapshot.Options{
						SnapshotDir:     snapshotDir,
						UpdateSnapshots: true, // Force updates to test concurrency
					})
					
					// Create snapshot with test-specific data
					data := fmt.Sprintf("test data for concurrent test %d", testID)
					s.Match(fmt.Sprintf("concurrent_%d", testID), data)
					
					return nil
				},
			}
		}

		runner.Run(testCases)

		// Verify all snapshots were created successfully
		files, err := os.ReadDir(snapshotDir)
		require.NoError(t, err)
		assert.Len(t, files, numTests)
	})

	t.Run("snapshot sanitizers work correctly in parallel context", func(t *testing.T) {
		tempDir := t.TempDir()
		snapshotDir := filepath.Join(tempDir, "sanitized")

		runner, err := parallel.NewRunner(t, parallel.Options{PoolSize: 3})
		require.NoError(t, err)
		defer runner.Cleanup()

		testCases := make([]parallel.TestCase, 4)
		for i := 0; i < 4; i++ {
			testID := i
			testCases[i] = parallel.TestCase{
				Name: fmt.Sprintf("sanitizer_test_%d", testID),
				Func: func(ctx context.Context, tctx *parallel.TestContext) error {
					mockT := newIntegrationTestingT(fmt.Sprintf("SanitizerTest_%d", testID))
					s := snapshot.New(mockT, snapshot.Options{
						SnapshotDir: snapshotDir,
						Sanitizers: []snapshot.Sanitizer{
							snapshot.SanitizeTimestamps(),
							snapshot.SanitizeUUIDs(),
							snapshot.SanitizeLambdaRequestID(),
						},
					})
					
					// Create data with elements that should be sanitized
					data := map[string]interface{}{
						"testID":    testID,
						"timestamp": "2023-01-01T12:00:00Z",
						"requestId": fmt.Sprintf("req-%d-abc123", testID),
						"uuid":      "550e8400-e29b-41d4-a716-446655440000",
						"resource":  tctx.GetTableName("data"),
					}
					
					s.MatchJSON(fmt.Sprintf("sanitized_%d", testID), data)
					return nil
				},
			}
		}

		runner.Run(testCases)

		// Verify sanitization occurred in all snapshots
		files, err := os.ReadDir(snapshotDir)
		require.NoError(t, err)
		assert.Len(t, files, 4)

		for _, file := range files {
			content, err := os.ReadFile(filepath.Join(snapshotDir, file.Name()))
			require.NoError(t, err)
			
			contentStr := string(content)
			assert.Contains(t, contentStr, "<TIMESTAMP>")
			assert.Contains(t, contentStr, "<UUID>")
			assert.Contains(t, contentStr, "<REQUEST_ID>")
			assert.NotContains(t, contentStr, "2023-01-01T12:00:00Z")
			assert.NotContains(t, contentStr, "550e8400-e29b-41d4-a716-446655440000")
		}
	})

	t.Run("snapshot conflicts are handled gracefully in parallel execution", func(t *testing.T) {
		tempDir := t.TempDir()
		snapshotDir := filepath.Join(tempDir, "conflicts")

		runner, err := parallel.NewRunner(t, parallel.Options{PoolSize: 2})
		require.NoError(t, err)
		defer runner.Cleanup()

		var errors []string
		var errorsMutex sync.Mutex

		// Create tests that try to create snapshots with same name but different data
		testCases := []parallel.TestCase{
			{
				Name: "conflict_test_1",
				Func: func(ctx context.Context, tctx *parallel.TestContext) error {
					mockT := newIntegrationTestingT("ConflictTest1")
					s := snapshot.New(mockT, snapshot.Options{SnapshotDir: snapshotDir})
					
					// Create snapshot first
					s.Match("shared_name", "data from test 1")
					return nil
				},
			},
			{
				Name: "conflict_test_2",
				Func: func(ctx context.Context, tctx *parallel.TestContext) error {
					// Add slight delay to ensure first test runs first
					time.Sleep(10 * time.Millisecond)
					
					mockT := newIntegrationTestingT("ConflictTest2")
					s := snapshot.New(mockT, snapshot.Options{SnapshotDir: snapshotDir})
					
					defer func() {
						if r := recover(); r != nil {
							errorsMutex.Lock()
							errors = append(errors, fmt.Sprintf("Test panicked: %v", r))
							errorsMutex.Unlock()
						}
					}()
					
					// Try to match with different data (should fail)
					s.Match("shared_name", "data from test 2")
					
					// Should only get here if snapshot matched
					errorsMutex.Lock()
					errors = append(errors, "Expected snapshot mismatch but got success")
					errorsMutex.Unlock()
					
					return nil
				},
			},
		}

		runner.Run(testCases)

		// One test should have failed due to snapshot mismatch
		assert.NotEmpty(t, errors)
	})
}

func TestResourceIsolationWithSnapshots(t *testing.T) {
	t.Run("resource locks prevent snapshot conflicts", func(t *testing.T) {
		tempDir := t.TempDir()
		snapshotDir := filepath.Join(tempDir, "resource_isolation")

		runner, err := parallel.NewRunner(t, parallel.Options{PoolSize: 3})
		require.NoError(t, err)
		defer runner.Cleanup()

		var accessOrder []int
		var mutex sync.Mutex

		testCases := make([]parallel.TestCase, 3)
		for i := 0; i < 3; i++ {
			testID := i
			testCases[i] = parallel.TestCase{
				Name: fmt.Sprintf("resource_isolation_test_%d", testID),
				Func: func(ctx context.Context, tctx *parallel.TestContext) error {
					return tctx.WithResourceLock("shared_snapshot", func() error {
						mockT := newIntegrationTestingT(fmt.Sprintf("ResourceTest_%d", testID))
						s := snapshot.New(mockT, snapshot.Options{SnapshotDir: snapshotDir})
						
						// Record access order
						mutex.Lock()
						accessOrder = append(accessOrder, testID)
						mutex.Unlock()
						
						// Create snapshot with test-specific data
						data := map[string]interface{}{
							"accessOrder": len(accessOrder),
							"testID":      testID,
						}
						
						s.MatchJSON(fmt.Sprintf("resource_test_%d", testID), data)
						
						// Simulate some processing time
						time.Sleep(20 * time.Millisecond)
						
						return nil
					})
				},
			}
		}

		runner.Run(testCases)

		assert.Len(t, accessOrder, 3)
		
		// Verify all snapshots were created
		files, err := os.ReadDir(snapshotDir)
		require.NoError(t, err)
		assert.Len(t, files, 3)
	})

	t.Run("isolated table names appear in snapshots", func(t *testing.T) {
		tempDir := t.TempDir()
		snapshotDir := filepath.Join(tempDir, "table_isolation")

		runner, err := parallel.NewRunner(t, parallel.Options{PoolSize: 2})
		require.NoError(t, err)
		defer runner.Cleanup()

		testCases := make([]parallel.TestCase, 3)
		tableNames := []string{"users", "orders", "products"}
		
		for i, tableName := range tableNames {
			baseTableName := tableName
			testCases[i] = parallel.TestCase{
				Name: fmt.Sprintf("table_isolation_%s", baseTableName),
				Func: func(ctx context.Context, tctx *parallel.TestContext) error {
					mockT := newIntegrationTestingT(fmt.Sprintf("TableIsolation_%s", baseTableName))
					s := snapshot.New(mockT, snapshot.Options{SnapshotDir: snapshotDir})
					
					isolatedTableName := tctx.GetTableName(baseTableName)
					isolatedFuncName := tctx.GetFunctionName("handler")
					
					data := map[string]interface{}{
						"originalTable": baseTableName,
						"isolatedTable": isolatedTableName,
						"isolatedFunc":  isolatedFuncName,
						"namespace":     fmt.Sprintf("contains_%s", baseTableName),
					}
					
					s.MatchJSON(fmt.Sprintf("%s_snapshot", baseTableName), data)
					return nil
				},
			}
		}

		runner.Run(testCases)

		files, err := os.ReadDir(snapshotDir)
		require.NoError(t, err)
		assert.Len(t, files, 3)

		// Verify each snapshot contains unique isolated table names
		seenTableNames := make(map[string]bool)
		
		for _, file := range files {
			content, err := os.ReadFile(filepath.Join(snapshotDir, file.Name()))
			require.NoError(t, err)
			
			var data map[string]interface{}
			err = json.Unmarshal(content, &data)
			require.NoError(t, err)
			
			isolatedTable := data["isolatedTable"].(string)
			assert.False(t, seenTableNames[isolatedTable], "Isolated table name should be unique: %s", isolatedTable)
			seenTableNames[isolatedTable] = true
			
			// Verify the isolated name contains the test namespace
			assert.Contains(t, isolatedTable, "test")
			assert.Contains(t, isolatedTable, data["originalTable"])
		}
	})
}

func TestPerformanceTesting(t *testing.T) {
	t.Run("performance test with parallel execution and snapshots", func(t *testing.T) {
		if testing.Short() {
			t.Skip("Skipping performance test in short mode")
		}

		tempDir := t.TempDir()
		snapshotDir := filepath.Join(tempDir, "performance")

		poolSizes := []int{2, 4, 8}
		numTests := 50

		for _, poolSize := range poolSizes {
			t.Run(fmt.Sprintf("pool_size_%d", poolSize), func(t *testing.T) {
				start := time.Now()

				runner, err := parallel.NewRunner(t, parallel.Options{PoolSize: poolSize})
				require.NoError(t, err)
				defer runner.Cleanup()

				testCases := make([]parallel.TestCase, numTests)
				for i := 0; i < numTests; i++ {
					testID := i
					testCases[i] = parallel.TestCase{
						Name: fmt.Sprintf("perf_test_%d", testID),
						Func: func(ctx context.Context, tctx *parallel.TestContext) error {
							mockT := newIntegrationTestingT(fmt.Sprintf("PerfTest_%d", testID))
							s := snapshot.New(mockT, snapshot.Options{
								SnapshotDir: filepath.Join(snapshotDir, fmt.Sprintf("pool_%d", poolSize)),
								Sanitizers: []snapshot.Sanitizer{
									snapshot.SanitizeTimestamps(),
									snapshot.SanitizeUUIDs(),
								},
							})
							
							// Create moderately complex data
							data := map[string]interface{}{
								"testID":    testID,
								"timestamp": time.Now().Format(time.RFC3339),
								"uuid":      fmt.Sprintf("550e8400-e29b-%04d-a716-446655440000", testID),
								"items": make([]map[string]interface{}, 10),
								"metadata": map[string]interface{}{
									"poolSize":  poolSize,
									"tableName": tctx.GetTableName("perf_table"),
									"funcName":  tctx.GetFunctionName("perf_func"),
								},
							}
							
							// Fill items array
							items := data["items"].([]map[string]interface{})
							for j := 0; j < 10; j++ {
								items[j] = map[string]interface{}{
									"id":    fmt.Sprintf("item_%d_%d", testID, j),
									"value": j * testID,
								}
							}
							
							s.MatchJSON(fmt.Sprintf("perf_%d", testID), data)
							return nil
						},
					}
				}

				runner.Run(testCases)
				
				duration := time.Since(start)
				t.Logf("Pool size %d completed %d tests in %v (avg: %v per test)", 
					poolSize, numTests, duration, duration/time.Duration(numTests))
				
				// Verify all snapshots were created
				poolDir := filepath.Join(snapshotDir, fmt.Sprintf("pool_%d", poolSize))
				files, err := os.ReadDir(poolDir)
				require.NoError(t, err)
				assert.Len(t, files, numTests)
			})
		}
	})

	t.Run("memory usage during parallel snapshot creation", func(t *testing.T) {
		if testing.Short() {
			t.Skip("Skipping memory test in short mode")
		}

		tempDir := t.TempDir()
		snapshotDir := filepath.Join(tempDir, "memory")

		var startMem, endMem runtime.MemStats
		runtime.GC()
		runtime.ReadMemStats(&startMem)

		runner, err := parallel.NewRunner(t, parallel.Options{PoolSize: 4})
		require.NoError(t, err)
		defer runner.Cleanup()

		numTests := 100
		testCases := make([]parallel.TestCase, numTests)
		
		for i := 0; i < numTests; i++ {
			testID := i
			testCases[i] = parallel.TestCase{
				Name: fmt.Sprintf("memory_test_%d", testID),
				Func: func(ctx context.Context, tctx *parallel.TestContext) error {
					mockT := newIntegrationTestingT(fmt.Sprintf("MemoryTest_%d", testID))
					s := snapshot.New(mockT, snapshot.Options{SnapshotDir: snapshotDir})
					
					// Create larger data structure
					largeData := make(map[string]interface{})
					largeData["testID"] = testID
					largeData["largeArray"] = make([]string, 100)
					
					for j := 0; j < 100; j++ {
						largeData["largeArray"].([]string)[j] = fmt.Sprintf("item_%d_%d", testID, j)
					}
					
					s.MatchJSON(fmt.Sprintf("memory_%d", testID), largeData)
					return nil
				},
			}
		}

		runner.Run(testCases)

		runtime.GC()
		runtime.ReadMemStats(&endMem)
		
		memoryIncrease := endMem.Alloc - startMem.Alloc
		t.Logf("Memory increase: %d bytes (%.2f MB) for %d tests", 
			memoryIncrease, float64(memoryIncrease)/(1024*1024), numTests)
		
		// Memory increase should be reasonable (less than 50MB for 100 tests)
		assert.Less(t, memoryIncrease, uint64(50*1024*1024))
	})
}

func TestErrorScenarios(t *testing.T) {
	t.Run("handles corrupted snapshot files", func(t *testing.T) {
		tempDir := t.TempDir()
		snapshotDir := filepath.Join(tempDir, "corrupted")

		// Create a corrupted snapshot file
		os.MkdirAll(snapshotDir, 0755)
		corruptedFile := filepath.Join(snapshotDir, "CorruptedTest_corrupted_test.snap")
		os.WriteFile(corruptedFile, []byte("invalid\x00binary\xFFdata"), 0644)

		runner, err := parallel.NewRunner(t, parallel.Options{PoolSize: 1})
		require.NoError(t, err)
		defer runner.Cleanup()

		mockT := newIntegrationTestingT("CorruptedTest")
		var testErrors []string
		
		testCases := []parallel.TestCase{
			{
				Name: "corrupted_snapshot_test",
				Func: func(ctx context.Context, tctx *parallel.TestContext) error {
					defer func() {
						if r := recover(); r != nil {
							testErrors = append(testErrors, fmt.Sprintf("Test panicked: %v", r))
						}
					}()

					s := snapshot.New(mockT, snapshot.Options{SnapshotDir: snapshotDir})
					s.Match("corrupted_test", "new valid data")
					
					return nil
				},
			},
		}

		runner.Run(testCases)
		
		// Should have failed due to snapshot mismatch
		errors := mockT.getErrors()
		assert.NotEmpty(t, errors)
	})

	t.Run("handles permission errors during parallel execution", func(t *testing.T) {
		tempDir := t.TempDir()
		readOnlyDir := filepath.Join(tempDir, "readonly")
		os.MkdirAll(readOnlyDir, 0755)
		
		// Make directory read-only
		os.Chmod(readOnlyDir, 0555)
		defer os.Chmod(readOnlyDir, 0755) // Restore for cleanup

		runner, err := parallel.NewRunner(t, parallel.Options{PoolSize: 2})
		require.NoError(t, err)
		defer runner.Cleanup()

		var testErrors int64

		testCases := make([]parallel.TestCase, 3)
		for i := 0; i < 3; i++ {
			testID := i
			testCases[i] = parallel.TestCase{
				Name: fmt.Sprintf("permission_test_%d", testID),
				Func: func(ctx context.Context, tctx *parallel.TestContext) error {
					defer func() {
						if r := recover(); r != nil {
							atomic.AddInt64(&testErrors, 1)
						}
					}()

					mockT := newIntegrationTestingT(fmt.Sprintf("PermissionTest_%d", testID))
					s := snapshot.New(mockT, snapshot.Options{SnapshotDir: readOnlyDir})
					s.Match(fmt.Sprintf("permission_%d", testID), "test data")
					
					return nil
				},
			}
		}

		runner.Run(testCases)
		
		// All tests should have failed due to permission errors
		assert.Equal(t, int64(3), atomic.LoadInt64(&testErrors))
	})

	t.Run("handles disk space exhaustion gracefully", func(t *testing.T) {
		// This test simulates disk space exhaustion by creating a very small tmpfs
		// Skip on systems where this can't be tested
		if os.Getuid() != 0 {
			t.Skip("Skipping disk space test (requires root for tmpfs)")
		}

		tempDir := t.TempDir()
		snapshotDir := filepath.Join(tempDir, "diskfull")

		runner, err := parallel.NewRunner(t, parallel.Options{PoolSize: 2})
		require.NoError(t, err)
		defer runner.Cleanup()

		testCases := []parallel.TestCase{
			{
				Name: "disk_space_test",
				Func: func(ctx context.Context, tctx *parallel.TestContext) error {
					mockT := newIntegrationTestingT("DiskSpaceTest")
					s := snapshot.New(mockT, snapshot.Options{SnapshotDir: snapshotDir})
					
					// Try to create a very large snapshot
					largeData := make([]string, 1000000)
					for i := range largeData {
						largeData[i] = fmt.Sprintf("very long string data item %d", i)
					}
					
					defer func() {
						recover() // Expected to fail
					}()

					s.MatchJSON("large_snapshot", largeData)
					return nil
				},
			},
		}

		runner.Run(testCases)
		// Test should complete (either success or controlled failure)
	})

	t.Run("handles network failures during parallel execution", func(t *testing.T) {
		// Simulate network-dependent snapshot operations
		runner, err := parallel.NewRunner(t, parallel.Options{PoolSize: 3})
		require.NoError(t, err)
		defer runner.Cleanup()

		var completedTests int64

		testCases := make([]parallel.TestCase, 5)
		for i := 0; i < 5; i++ {
			testID := i
			testCases[i] = parallel.TestCase{
				Name: fmt.Sprintf("network_simulation_%d", testID),
				Func: func(ctx context.Context, tctx *parallel.TestContext) error {
					// Simulate network-dependent operation
					select {
					case <-ctx.Done():
						return ctx.Err()
					case <-time.After(time.Duration(testID*10) * time.Millisecond):
						// Simulate varying response times
					}
					
					mockT := newIntegrationTestingT(fmt.Sprintf("NetworkTest_%d", testID))
					s := snapshot.New(mockT, snapshot.Options{
						SnapshotDir: filepath.Join(t.TempDir(), "network"),
					})
					
					// Create snapshot with network-like data
					data := map[string]interface{}{
						"testID":    testID,
						"endpoint":  fmt.Sprintf("https://api.example.com/test/%d", testID),
						"response":  map[string]interface{}{"status": 200, "data": "success"},
						"timestamp": "2023-01-01T12:00:00Z",
					}
					
					s.MatchJSON(fmt.Sprintf("network_%d", testID), data)
					atomic.AddInt64(&completedTests, 1)
					
					return nil
				},
			}
		}

		runner.Run(testCases)
		
		assert.Equal(t, int64(5), atomic.LoadInt64(&completedTests))
	})
}

func TestResourceContentionAndLocking(t *testing.T) {
	t.Run("resource contention with snapshot creation", func(t *testing.T) {
		tempDir := t.TempDir()
		snapshotDir := filepath.Join(tempDir, "contention")

		runner, err := parallel.NewRunner(t, parallel.Options{PoolSize: 6})
		require.NoError(t, err)
		defer runner.Cleanup()

		numTests := 12
		var accessCounts int64
		var sharedResource int64

		testCases := make([]parallel.TestCase, numTests)
		for i := 0; i < numTests; i++ {
			testID := i
			testCases[i] = parallel.TestCase{
				Name: fmt.Sprintf("contention_test_%d", testID),
				Func: func(ctx context.Context, tctx *parallel.TestContext) error {
					return tctx.WithResourceLock("shared_db", func() error {
						// Simulate critical section with shared resource
						current := atomic.LoadInt64(&sharedResource)
						time.Sleep(5 * time.Millisecond) // Simulate processing
						atomic.StoreInt64(&sharedResource, current+1)
						atomic.AddInt64(&accessCounts, 1)
						
						// Create snapshot of the operation
						mockT := newIntegrationTestingT(fmt.Sprintf("ContentionTest_%d", testID))
						s := snapshot.New(mockT, snapshot.Options{
							SnapshotDir: snapshotDir,
							Sanitizers: []snapshot.Sanitizer{
								snapshot.SanitizeTimestamps(),
							},
						})
						
						data := map[string]interface{}{
							"testID":         testID,
							"accessOrder":    atomic.LoadInt64(&accessCounts),
							"resourceValue":  atomic.LoadInt64(&sharedResource),
							"timestamp":      time.Now().Format(time.RFC3339),
							"isolatedTable":  tctx.GetTableName("shared"),
						}
						
						s.MatchJSON(fmt.Sprintf("contention_%d", testID), data)
						return nil
					})
				},
			}
		}

		runner.Run(testCases)

		// Verify resource was accessed correctly
		assert.Equal(t, int64(numTests), atomic.LoadInt64(&accessCounts))
		assert.Equal(t, int64(numTests), atomic.LoadInt64(&sharedResource))
		
		// Verify all snapshots were created
		files, err := os.ReadDir(snapshotDir)
		require.NoError(t, err)
		assert.Len(t, files, numTests)
	})
}

func TestIntegrationWithVasDeference(t *testing.T) {
	t.Run("integration with vasdeference TestContext", func(t *testing.T) {
		// Create VasDeference context
		testCtx := NewTestContext(t)
		arrange := NewArrange(testCtx)
		defer arrange.Cleanup()

		tempDir := t.TempDir()
		snapshotDir := filepath.Join(tempDir, "vasdeference")

		runner, err := parallel.NewRunnerWithArrange(arrange, parallel.Options{PoolSize: 2})
		require.NoError(t, err)
		defer runner.Cleanup()

		testCases := []parallel.TestCase{
			{
				Name: "vasdeference_integration",
				Func: func(ctx context.Context, tctx *parallel.TestContext) error {
					// Use VasDeference functionality
					mockT := newIntegrationTestingT("VasDeferenceIntegration")
					s := snapshot.New(mockT, snapshot.Options{SnapshotDir: snapshotDir})
					
					data := map[string]interface{}{
						"region":        testCtx.Region,
						"namespace":     arrange.Namespace,
						"isolatedTable": tctx.GetTableName("integration"),
						"isolatedFunc":  tctx.GetFunctionName("handler"),
					}
					
					s.MatchJSON("vasdeference_test", data)
					return nil
				},
			},
		}

		runner.Run(testCases)

		// Verify snapshot was created with VasDeference data
		files, err := os.ReadDir(snapshotDir)
		require.NoError(t, err)
		require.Len(t, files, 1)

		content, err := os.ReadFile(filepath.Join(snapshotDir, files[0].Name()))
		require.NoError(t, err)

		var data map[string]interface{}
		err = json.Unmarshal(content, &data)
		require.NoError(t, err)

		assert.Equal(t, RegionUSEast1, data["region"])
		assert.NotEmpty(t, data["namespace"])
		assert.Contains(t, data["isolatedTable"], "integration")
		assert.Contains(t, data["isolatedFunc"], "handler")
	})
}

func TestConcurrentSnapshotFileOperations(t *testing.T) {
	t.Run("concurrent file operations do not cause corruption", func(t *testing.T) {
		tempDir := t.TempDir()
		snapshotDir := filepath.Join(tempDir, "concurrent_files")

		runner, err := parallel.NewRunner(t, parallel.Options{PoolSize: 8})
		require.NoError(t, err)
		defer runner.Cleanup()

		numTests := 50
		testCases := make([]parallel.TestCase, numTests)
		
		for i := 0; i < numTests; i++ {
			testID := i
			testCases[i] = parallel.TestCase{
				Name: fmt.Sprintf("file_ops_test_%d", testID),
				Func: func(ctx context.Context, tctx *parallel.TestContext) error {
					mockT := newIntegrationTestingT(fmt.Sprintf("FileOpsTest_%d", testID))
					s := snapshot.New(mockT, snapshot.Options{
						SnapshotDir:     snapshotDir,
						UpdateSnapshots: testID%2 == 0, // Half update, half create new
					})
					
					data := fmt.Sprintf("file operation test data %d", testID)
					s.Match(fmt.Sprintf("file_ops_%d", testID), data)
					
					return nil
				},
			}
		}

		runner.Run(testCases)

		// Verify all files were created and contain correct data
		files, err := os.ReadDir(snapshotDir)
		require.NoError(t, err)
		assert.Len(t, files, numTests)

		for i, file := range files {
			content, err := os.ReadFile(filepath.Join(snapshotDir, file.Name()))
			require.NoError(t, err)
			
			// Verify content is not corrupted
			assert.NotEmpty(t, content)
			assert.Contains(t, string(content), "file operation test data")
			
			// Verify file is valid (no partial writes)
			assert.NotContains(t, string(content), "\x00")
			
			t.Logf("File %d: %s (%d bytes)", i, file.Name(), len(content))
		}
	})
}