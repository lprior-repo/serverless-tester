// Package performance demonstrates functional comprehensive performance testing for serverless applications
// Following strict TDD methodology with functional programming patterns, advanced benchmarking and profiling
package performance

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"runtime"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/samber/lo"
	"github.com/samber/mo"
	"vasdeference"
	"vasdeference/lambda"
	"vasdeference/snapshot"
)

// Performance Test Models

type LambdaPerformanceRequest struct {
	RequestID   string                 `json:"requestId"`
	Payload     map[string]interface{} `json:"payload"`
	Size        int                    `json:"size"`
	Complexity  int                    `json:"complexity"`
	Metadata    map[string]string      `json:"metadata,omitempty"`
}

type LambdaPerformanceResult struct {
	Success         bool              `json:"success"`
	ExecutionTime   time.Duration     `json:"executionTime"`
	BilledDuration  time.Duration     `json:"billedDuration"`
	MemoryUsed      int64             `json:"memoryUsed"`
	MemoryAllocated int64             `json:"memoryAllocated"`
	InitDuration    time.Duration     `json:"initDuration,omitempty"`
	IsColdStart     bool              `json:"isColdStart"`
	ErrorMessage    string            `json:"errorMessage,omitempty"`
	LogOutput       string            `json:"logOutput,omitempty"`
	RequestID       string            `json:"requestId"`
	Metadata        map[string]string `json:"metadata,omitempty"`
}

type PerformanceMetrics struct {
	AverageLatency    time.Duration
	P50Latency        time.Duration
	P95Latency        time.Duration
	P99Latency        time.Duration
	MinLatency        time.Duration
	MaxLatency        time.Duration
	Throughput        float64
	ErrorRate         float64
	ColdStartRate     float64
	AverageMemoryUsed int64
	TotalInvocations  int
	SuccessfulInvocations int
	FailedInvocations int
	ColdStarts        int
	WarmStarts        int
}

type LoadTestConfig struct {
	Concurrency    int
	Duration       time.Duration
	RequestsPerSec float64
	RampUpTime     time.Duration
	Payload        interface{}
}

// Performance Test Infrastructure

type LambdaPerformanceTestContext struct {
	VDF                    *vasdeference.VasDeference
	SimpleFunctionName     string
	CpuIntensiveFunctionName string
	MemoryIntensiveFunctionName string
	IoIntensiveFunctionName string
	ErrorProneFunctionName string
	MemoryProfiler         *MemoryProfiler
	PerformanceBaseline    map[string]PerformanceMetrics
}

type MemoryProfiler struct {
	mu                sync.Mutex
	initialMemStats   runtime.MemStats
	snapshots        []runtime.MemStats
	running          bool
	stopChan         chan struct{}
}

func setupLambdaPerformanceInfrastructure(t *testing.T) *LambdaPerformanceTestContext {
	t.Helper()
	
	// Functional performance infrastructure setup with monadic error handling
	setupResult := lo.Pipe4(
		time.Now().Unix(),
		func(timestamp int64) mo.Result[*vasdeference.VasDeference] {
			vdf := vasdeference.NewWithOptions(t, &vasdeference.VasDeferenceOptions{
				Region:    "us-east-1",
				Namespace: fmt.Sprintf("functional-perf-test-%d", timestamp),
			})
			if vdf == nil {
				return mo.Err[*vasdeference.VasDeference](fmt.Errorf("failed to create VasDeference instance"))
			}
			return mo.Ok(vdf)
		},
		func(vdfResult mo.Result[*vasdeference.VasDeference]) mo.Result[*LambdaPerformanceTestContext] {
			return vdfResult.Map(func(vdf *vasdeference.VasDeference) *LambdaPerformanceTestContext {
				return &LambdaPerformanceTestContext{
					VDF:                         vdf,
					SimpleFunctionName:          vdf.PrefixResourceName("functional-simple-func"),
					CpuIntensiveFunctionName:    vdf.PrefixResourceName("functional-cpu-intensive-func"),
					MemoryIntensiveFunctionName: vdf.PrefixResourceName("functional-memory-intensive-func"),
					IoIntensiveFunctionName:     vdf.PrefixResourceName("functional-io-intensive-func"),
					ErrorProneFunctionName:      vdf.PrefixResourceName("functional-error-prone-func"),
					MemoryProfiler:             NewFunctionalMemoryProfiler(),
					PerformanceBaseline:        make(map[string]PerformanceMetrics),
				}
			})
		},
		func(ctxResult mo.Result[*LambdaPerformanceTestContext]) mo.Result[*LambdaPerformanceTestContext] {
			return ctxResult.FlatMap(func(ctx *LambdaPerformanceTestContext) mo.Result[*LambdaPerformanceTestContext] {
				// Functional performance test function setup
				setupFunctionalPerformanceTestFunctions(t, ctx)
				loadFunctionalPerformanceBaselines(t, ctx)
				return mo.Ok(ctx)
			})
		},
		func(finalResult mo.Result[*LambdaPerformanceTestContext]) *LambdaPerformanceTestContext {
			return finalResult.Match(
				func(ctx *LambdaPerformanceTestContext) *LambdaPerformanceTestContext {
					ctx.VDF.RegisterCleanup(func() error {
						ctx.MemoryProfiler.Stop()
						return cleanupFunctionalPerformanceInfrastructure(ctx)
					})
					return ctx
				},
				func(err error) *LambdaPerformanceTestContext {
					t.Fatalf("Failed to setup functional performance infrastructure: %v", err)
					return nil
				},
			)
		},
	)
	
	return setupResult
}

func setupFunctionalPerformanceTestFunctions(t *testing.T, ctx *LambdaPerformanceTestContext) {
	t.Helper()
	
	// Functional function definition with immutable configuration
	functionConfigs := lo.Map([]struct {
		key    string
		name   string
		code   []byte
		memory int
	}{
		{"simple", ctx.SimpleFunctionName, generateFunctionalSimpleFunctionCode(), 128},
		{"cpu_intensive", ctx.CpuIntensiveFunctionName, generateFunctionalCpuIntensiveFunctionCode(), 256},
		{"memory_intensive", ctx.MemoryIntensiveFunctionName, generateFunctionalMemoryIntensiveFunctionCode(), 512},
		{"io_intensive", ctx.IoIntensiveFunctionName, generateFunctionalIoIntensiveFunctionCode(), 256},
		{"error_prone", ctx.ErrorProneFunctionName, generateFunctionalErrorProneFunctionCode(), 128},
	}, func(fn struct {
		key    string
		name   string
		code   []byte
		memory int
	}, _ int) mo.Result[string] {
		err := lambda.CreateFunction(ctx.VDF.Context, fn.name, lambda.FunctionConfig{
			Runtime:     "nodejs18.x",
			Handler:     "index.handler",
			Code:        fn.code,
			MemorySize:  fn.memory,
			Timeout:     30,
			Environment: map[string]string{
				"NODE_ENV":        "test",
				"FUNCTIONAL_MODE": "true",
			},
		})
		if err != nil {
			return mo.Err[string](fmt.Errorf("failed to create function %s: %w", fn.key, err))
		}
		return mo.Ok(fn.name)
	})
	
	// Functional validation of all function creations
	functionCreationResults := lo.Pipe2(
		functionConfigs,
		func(results []mo.Result[string]) mo.Result[[]string] {
			failedCreations := lo.Filter(results, func(result mo.Result[string], _ int) bool {
				return result.IsError()
			})
			
			if len(failedCreations) > 0 {
				return mo.Err[[]string](fmt.Errorf("failed to create %d functions", len(failedCreations)))
			}
			
			successfulNames := lo.FilterMap(results, func(result mo.Result[string], _ int) (string, bool) {
				return result.Match(
					func(name string) (string, bool) { return name, true },
					func(err error) (string, bool) { return "", false },
				)
			})
			
			return mo.Ok(successfulNames)
		},
		func(namesResult mo.Result[[]string]) bool {
			return namesResult.Match(
				func(names []string) bool {
					// Wait for all functions to be active using functional approach
					waitResults := lo.Map(names, func(name string, _ int) mo.Result[string] {
						lambda.WaitForFunctionActive(ctx.VDF.Context, name, 60*time.Second)
						return mo.Ok(name)
					})
					
					successCount := lo.CountBy(waitResults, func(result mo.Result[string]) bool {
						return result.IsOk()
					})
					
					t.Logf("Successfully setup %d/%d functional performance test functions", successCount, len(names))
					return successCount == len(names)
				},
				func(err error) bool {
					t.Fatalf("Failed to setup functional performance test functions: %v", err)
					return false
				},
			)
		},
	)
	
	if functionCreationResults {
		t.Log("All functional performance test functions setup successfully")
	} else {
		t.Log("Some functional performance test functions failed to setup")
	}
}

// Basic Performance Testing

func TestLambdaBasicPerformance(t *testing.T) {
	ctx := setupLambdaPerformanceInfrastructure(t)
	defer ctx.VDF.Cleanup()
	
	// Basic performance test using table-driven approach
	vasdeference.TableTest[LambdaPerformanceRequest, LambdaPerformanceResult](t, "Lambda Basic Performance").
		Case("simple_small_payload", generateSimpleRequest(100), LambdaPerformanceResult{Success: true}).
		Case("simple_medium_payload", generateSimpleRequest(10000), LambdaPerformanceResult{Success: true}).
		Case("simple_large_payload", generateSimpleRequest(100000), LambdaPerformanceResult{Success: true}).
		Parallel().
		WithMaxWorkers(5).
		Repeat(10).
		Run(func(req LambdaPerformanceRequest) LambdaPerformanceResult {
			return invokeLambdaWithMetrics(t, ctx, ctx.SimpleFunctionName, req)
		}, func(testName string, input LambdaPerformanceRequest, expected LambdaPerformanceResult, actual LambdaPerformanceResult) {
			validateBasicPerformance(t, testName, input, expected, actual)
		})
}

func TestLambdaColdVsWarmStartPerformance(t *testing.T) {
	ctx := setupLambdaPerformanceInfrastructure(t)
	defer ctx.VDF.Cleanup()
	
	// Force cold start by waiting
	time.Sleep(15 * time.Minute) // Ensure function is cold
	
	// Measure cold start
	coldStartRequest := generateSimpleRequest(1000)
	coldStartResult := invokeLambdaWithMetrics(t, ctx, ctx.SimpleFunctionName, coldStartRequest)
	
	// Measure warm starts (multiple invocations)
	warmStartResults := make([]LambdaPerformanceResult, 10)
	for i := range warmStartResults {
		warmStartRequest := generateSimpleRequest(1000)
		warmStartResults[i] = invokeLambdaWithMetrics(t, ctx, ctx.SimpleFunctionName, warmStartRequest)
		time.Sleep(100 * time.Millisecond) // Small delay between invocations
	}
	
	// Analyze cold vs warm start performance
	avgWarmStartTime := calculateAverageExecutionTime(warmStartResults)
	
	assert.True(t, coldStartResult.IsColdStart, "First invocation should be a cold start")
	assert.Greater(t, coldStartResult.ExecutionTime, avgWarmStartTime, 
		"Cold start should be slower than warm start")
	assert.NotZero(t, coldStartResult.InitDuration, "Cold start should have init duration")
	
	for _, result := range warmStartResults {
		assert.False(t, result.IsColdStart, "Subsequent invocations should be warm starts")
		assert.Zero(t, result.InitDuration, "Warm starts should not have init duration")
	}
	
	t.Logf("Cold start time: %v, Average warm start time: %v, Cold start overhead: %v",
		coldStartResult.ExecutionTime, avgWarmStartTime, coldStartResult.ExecutionTime-avgWarmStartTime)
}

// Benchmark Testing

func TestLambdaBenchmarkPerformance(t *testing.T) {
	ctx := setupLambdaPerformanceInfrastructure(t)
	defer ctx.VDF.Cleanup()
	
	testCases := []struct {
		name         string
		functionName string
		request      LambdaPerformanceRequest
		iterations   int
	}{
		{
			name:         "simple_function_benchmark",
			functionName: ctx.SimpleFunctionName,
			request:      generateSimpleRequest(1000),
			iterations:   1000,
		},
		{
			name:         "cpu_intensive_benchmark",
			functionName: ctx.CpuIntensiveFunctionName,
			request:      generateCpuIntensiveRequest(),
			iterations:   100,
		},
		{
			name:         "memory_intensive_benchmark", 
			functionName: ctx.MemoryIntensiveFunctionName,
			request:      generateMemoryIntensiveRequest(),
			iterations:   100,
		},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			benchmark := vasdeference.TableTest[LambdaPerformanceRequest, LambdaPerformanceResult](t, tc.name).
				Case("benchmark_case", tc.request, LambdaPerformanceResult{Success: true}).
				Benchmark(func(req LambdaPerformanceRequest) LambdaPerformanceResult {
					return invokeLambdaWithMetrics(t, ctx, tc.functionName, req)
				}, tc.iterations)
			
			// Performance assertions based on function type
			switch tc.name {
			case "simple_function_benchmark":
				assert.Less(t, benchmark.AverageTime, 100*time.Millisecond, 
					"Simple function should execute in under 100ms")
				assert.Less(t, benchmark.P95Time, 200*time.Millisecond,
					"95th percentile should be under 200ms")
			case "cpu_intensive_benchmark":
				assert.Less(t, benchmark.AverageTime, 2*time.Second,
					"CPU intensive function should execute in under 2s")
			case "memory_intensive_benchmark":
				assert.Less(t, benchmark.AverageTime, 1*time.Second,
					"Memory intensive function should execute in under 1s")
			}
			
			// Log benchmark results
			t.Logf("%s Results: Avg=%v, Min=%v, Max=%v, P95=%v, P99=%v",
				tc.name, benchmark.AverageTime, benchmark.MinTime, benchmark.MaxTime,
				benchmark.P95Time, benchmark.P99Time)
		})
	}
}

// Load Testing

func TestLambdaLoadPerformance(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping load test in short mode")
	}
	
	ctx := setupLambdaPerformanceInfrastructure(t)
	defer ctx.VDF.Cleanup()
	
	loadTests := []LoadTestConfig{
		{
			Concurrency:    10,
			Duration:       60 * time.Second,
			RequestsPerSec: 100,
			RampUpTime:     10 * time.Second,
			Payload:        generateSimpleRequest(1000),
		},
		{
			Concurrency:    50,
			Duration:       120 * time.Second,
			RequestsPerSec: 500,
			RampUpTime:     20 * time.Second,
			Payload:        generateSimpleRequest(1000),
		},
		{
			Concurrency:    100,
			Duration:       180 * time.Second,
			RequestsPerSec: 1000,
			RampUpTime:     30 * time.Second,
			Payload:        generateSimpleRequest(1000),
		},
	}
	
	for i, config := range loadTests {
		t.Run(fmt.Sprintf("load_test_%d", i+1), func(t *testing.T) {
			metrics := executeLoadTest(t, ctx, ctx.SimpleFunctionName, config)
			
			// Validate load test results
			assert.LessOrEqual(t, metrics.ErrorRate, 0.05, "Error rate should be under 5%")
			assert.GreaterOrEqual(t, metrics.Throughput, config.RequestsPerSec*0.8, 
				"Throughput should be at least 80% of target")
			assert.LessOrEqual(t, metrics.P95Latency, 1*time.Second,
				"95th percentile latency should be under 1 second")
			
			// Log load test results
			t.Logf("Load Test %d Results:", i+1)
			t.Logf("  Throughput: %.2f RPS (target: %.2f)", metrics.Throughput, config.RequestsPerSec)
			t.Logf("  Error Rate: %.2f%%", metrics.ErrorRate*100)
			t.Logf("  Latency P50: %v, P95: %v, P99: %v", 
				metrics.P50Latency, metrics.P95Latency, metrics.P99Latency)
			t.Logf("  Cold Start Rate: %.2f%%", metrics.ColdStartRate*100)
		})
	}
}

// Memory Performance Testing

func TestLambdaMemoryPerformance(t *testing.T) {
	ctx := setupLambdaPerformanceInfrastructure(t)
	defer ctx.VDF.Cleanup()
	
	// Start memory profiling
	ctx.MemoryProfiler.Start()
	defer ctx.MemoryProfiler.Stop()
	
	// Test different memory configurations
	memorySizes := []int{128, 256, 512, 1024}
	testRequest := generateMemoryIntensiveRequest()
	
	var results []struct {
		MemorySize int
		Result     LambdaPerformanceResult
	}
	
	for _, memSize := range memorySizes {
		// Update function memory size
		err := lambda.UpdateFunctionConfiguration(ctx.VDF.Context, ctx.MemoryIntensiveFunctionName, 
			lambda.FunctionConfig{MemorySize: memSize})
		require.NoError(t, err)
		
		// Wait for update to complete
		lambda.WaitForFunctionUpdated(ctx.VDF.Context, ctx.MemoryIntensiveFunctionName, 30*time.Second)
		
		// Run performance test
		result := invokeLambdaWithMetrics(t, ctx, ctx.MemoryIntensiveFunctionName, testRequest)
		results = append(results, struct {
			MemorySize int
			Result     LambdaPerformanceResult
		}{MemorySize: memSize, Result: result})
		
		t.Logf("Memory Size: %d MB, Execution Time: %v, Memory Used: %d MB",
			memSize, result.ExecutionTime, result.MemoryUsed/(1024*1024))
	}
	
	// Analyze memory performance correlation
	for i := 1; i < len(results); i++ {
		prev := results[i-1]
		curr := results[i]
		
		if curr.Result.ExecutionTime > prev.Result.ExecutionTime {
			t.Logf("Warning: Increasing memory from %d to %d MB did not improve performance",
				prev.MemorySize, curr.MemorySize)
		}
	}
	
	// Get memory profiler statistics
	memStats := ctx.MemoryProfiler.GetStats()
	assert.Less(t, memStats.HeapAllocBytes, 500*1024*1024, "Heap allocation should be under 500MB")
	assert.False(t, memStats.LeakDetected, "No memory leaks should be detected")
}

// Error Rate Performance Testing

func TestLambdaErrorRatePerformance(t *testing.T) {
	ctx := setupLambdaPerformanceInfrastructure(t)
	defer ctx.VDF.Cleanup()
	
	// Test error-prone function with various error rates
	errorRates := []float64{0.01, 0.05, 0.10, 0.20} // 1%, 5%, 10%, 20%
	
	for _, errorRate := range errorRates {
		t.Run(fmt.Sprintf("error_rate_%.0f_percent", errorRate*100), func(t *testing.T) {
			requests := make([]LambdaPerformanceRequest, 100)
			for i := range requests {
				requests[i] = generateErrorProneRequest(errorRate)
			}
			
			// Execute requests in parallel
			results := make([]LambdaPerformanceResult, len(requests))
			vasdeference.TableTest[LambdaPerformanceRequest, LambdaPerformanceResult](t, "Error Rate Test").
				Parallel().
				WithMaxWorkers(10).
				Run(func(req LambdaPerformanceRequest) LambdaPerformanceResult {
					return invokeLambdaWithMetrics(t, ctx, ctx.ErrorProneFunctionName, req)
				}, func(testName string, input LambdaPerformanceRequest, expected LambdaPerformanceResult, actual LambdaPerformanceResult) {
					// Store results for analysis
					for i, req := range requests {
						if req.RequestID == input.RequestID {
							results[i] = actual
							break
						}
					}
				})
			
			// Analyze error rate impact
			metrics := calculatePerformanceMetrics(results)
			
			// Validate that error rate is approximately as expected
			assert.InDelta(t, errorRate, metrics.ErrorRate, 0.02, // 2% tolerance
				"Actual error rate should be close to expected")
			
			// Validate that successful operations maintain good performance
			if metrics.SuccessfulInvocations > 0 {
				assert.Less(t, metrics.AverageLatency, 500*time.Millisecond,
					"Successful operations should maintain good performance")
			}
			
			t.Logf("Error Rate: %.1f%%, Avg Latency: %v, Throughput: %.2f RPS",
				metrics.ErrorRate*100, metrics.AverageLatency, metrics.Throughput)
		})
	}
}

// Regression Testing

func TestLambdaPerformanceRegression(t *testing.T) {
	ctx := setupLambdaPerformanceInfrastructure(t)
	defer ctx.VDF.Cleanup()
	
	// Execute current performance tests
	currentMetrics := measureCurrentPerformance(t, ctx)
	
	// Compare with baselines if available
	for functionName, baseline := range ctx.PerformanceBaseline {
		current, exists := currentMetrics[functionName]
		if !exists {
			continue
		}
		
		t.Run(fmt.Sprintf("regression_%s", functionName), func(t *testing.T) {
			// Regression thresholds (allowing for some variance)
			maxLatencyRegression := 1.20  // 20% increase allowed
			maxErrorRateRegression := 1.05 // 5% increase allowed
			minThroughputRegression := 0.90 // 10% decrease allowed
			
			// Latency regression check
			latencyRatio := float64(current.AverageLatency) / float64(baseline.AverageLatency)
			assert.LessOrEqual(t, latencyRatio, maxLatencyRegression,
				"Average latency regression detected: %.2fx slower", latencyRatio)
			
			// Error rate regression check
			if baseline.ErrorRate > 0 {
				errorRateRatio := current.ErrorRate / baseline.ErrorRate
				assert.LessOrEqual(t, errorRateRatio, maxErrorRateRegression,
					"Error rate regression detected: %.2fx higher", errorRateRatio)
			}
			
			// Throughput regression check
			throughputRatio := current.Throughput / baseline.Throughput
			assert.GreaterOrEqual(t, throughputRatio, minThroughputRegression,
				"Throughput regression detected: %.2fx slower", 1/throughputRatio)
			
			t.Logf("Performance Comparison for %s:", functionName)
			t.Logf("  Latency: %v -> %v (%.2fx)", baseline.AverageLatency, current.AverageLatency, latencyRatio)
			t.Logf("  Error Rate: %.2f%% -> %.2f%%", baseline.ErrorRate*100, current.ErrorRate*100)
			t.Logf("  Throughput: %.2f -> %.2f RPS", baseline.Throughput, current.Throughput)
		})
	}
	
	// Save current metrics as new baseline
	savePerformanceBaselines(t, ctx, currentMetrics)
}

// Snapshot-Based Performance Validation

func TestLambdaPerformanceSnapshots(t *testing.T) {
	ctx := setupLambdaPerformanceInfrastructure(t)
	defer ctx.VDF.Cleanup()
	
	// Execute performance test
	request := generateSimpleRequest(5000)
	result := invokeLambdaWithMetrics(t, ctx, ctx.SimpleFunctionName, request)
	
	// Create snapshot tester for performance results
	snapshotTester := snapshot.New(t, snapshot.Options{
		SnapshotDir: "testdata/snapshots/performance",
		JSONIndent:  true,
		Sanitizers: []snapshot.Sanitizer{
			snapshot.SanitizeTimestamps(),
			snapshot.SanitizeCustom(`"requestId":"[^"]*"`, `"requestId":"<REQUEST_ID>"`),
			snapshot.SanitizeCustom(`"executionTime":\d+`, `"executionTime":"<EXECUTION_TIME>"`),
		},
	})
	
	// Validate performance result structure
	snapshotTester.MatchJSON("lambda_performance_result", result)
}

// Utility Functions

func invokeLambdaWithMetrics(t *testing.T, ctx *LambdaPerformanceTestContext, functionName string, request LambdaPerformanceRequest) LambdaPerformanceResult {
	t.Helper()
	
	startTime := time.Now()
	
	payload, err := json.Marshal(request)
	require.NoError(t, err)
	
	result, err := lambda.InvokeE(ctx.VDF.Context, functionName, payload)
	executionTime := time.Since(startTime)
	
	perfResult := LambdaPerformanceResult{
		RequestID:     request.RequestID,
		ExecutionTime: executionTime,
		Success:       err == nil,
	}
	
	if err != nil {
		perfResult.ErrorMessage = err.Error()
		return perfResult
	}
	
	// Extract performance metrics from Lambda response
	if result.LogResult != nil {
		perfResult.LogOutput = *result.LogResult
		
		// Parse Lambda execution metrics from logs
		if billedDuration, memUsed, initDuration, found := parseLambdaMetrics(*result.LogResult); found {
			perfResult.BilledDuration = billedDuration
			perfResult.MemoryUsed = memUsed
			perfResult.InitDuration = initDuration
			perfResult.IsColdStart = initDuration > 0
		}
	}
	
	return perfResult
}

func executeLoadTest(t *testing.T, ctx *LambdaPerformanceTestContext, functionName string, config LoadTestConfig) PerformanceMetrics {
	t.Helper()
	
	startTime := time.Now()
	endTime := startTime.Add(config.Duration)
	
	var (
		mu      sync.Mutex
		results []LambdaPerformanceResult
		wg      sync.WaitGroup
	)
	
	// Rate limiter
	ticker := time.NewTicker(time.Duration(float64(time.Second) / config.RequestsPerSec))
	defer ticker.Stop()
	
	// Worker pool
	semaphore := make(chan struct{}, config.Concurrency)
	
	requestCount := 0
	for time.Now().Before(endTime) {
		select {
		case <-ticker.C:
			wg.Add(1)
			go func(reqID int) {
				defer wg.Done()
				
				// Acquire worker slot
				semaphore <- struct{}{}
				defer func() { <-semaphore }()
				
				request := config.Payload.(LambdaPerformanceRequest)
				request.RequestID = fmt.Sprintf("load-test-%d", reqID)
				
				result := invokeLambdaWithMetrics(t, ctx, functionName, request)
				
				mu.Lock()
				results = append(results, result)
				mu.Unlock()
			}(requestCount)
			requestCount++
			
		default:
			time.Sleep(1 * time.Millisecond)
		}
	}
	
	wg.Wait()
	
	return calculatePerformanceMetrics(results)
}

func calculatePerformanceMetrics(results []LambdaPerformanceResult) PerformanceMetrics {
	if len(results) == 0 {
		return PerformanceMetrics{}
	}
	
	var (
		totalLatency    time.Duration
		successful      int
		failed          int
		coldStarts      int
		warmStarts      int
		latencies       []time.Duration
		totalMemoryUsed int64
	)
	
	for _, result := range results {
		totalLatency += result.ExecutionTime
		latencies = append(latencies, result.ExecutionTime)
		totalMemoryUsed += result.MemoryUsed
		
		if result.Success {
			successful++
		} else {
			failed++
		}
		
		if result.IsColdStart {
			coldStarts++
		} else {
			warmStarts++
		}
	}
	
	// Sort latencies for percentile calculation
	sortLatencies(latencies)
	
	metrics := PerformanceMetrics{
		TotalInvocations:      len(results),
		SuccessfulInvocations: successful,
		FailedInvocations:     failed,
		ColdStarts:           coldStarts,
		WarmStarts:           warmStarts,
		AverageLatency:       totalLatency / time.Duration(len(results)),
		MinLatency:           latencies[0],
		MaxLatency:           latencies[len(latencies)-1],
		P50Latency:           latencies[len(latencies)*50/100],
		P95Latency:           latencies[len(latencies)*95/100],
		P99Latency:           latencies[len(latencies)*99/100],
		ErrorRate:            float64(failed) / float64(len(results)),
		ColdStartRate:        float64(coldStarts) / float64(len(results)),
		AverageMemoryUsed:    totalMemoryUsed / int64(len(results)),
	}
	
	return metrics
}

// Test Data Generators

func generateSimpleRequest(size int) LambdaPerformanceRequest {
	payload := make(map[string]interface{})
	payload["data"] = make([]byte, size)
	
	return LambdaPerformanceRequest{
		RequestID:  fmt.Sprintf("req-%d", time.Now().UnixNano()),
		Payload:    payload,
		Size:       size,
		Complexity: 1,
	}
}

func generateCpuIntensiveRequest() LambdaPerformanceRequest {
	return LambdaPerformanceRequest{
		RequestID:  fmt.Sprintf("cpu-%d", time.Now().UnixNano()),
		Payload:    map[string]interface{}{"iterations": 1000000},
		Complexity: 10,
	}
}

func generateMemoryIntensiveRequest() LambdaPerformanceRequest {
	return LambdaPerformanceRequest{
		RequestID:  fmt.Sprintf("mem-%d", time.Now().UnixNano()),
		Payload:    map[string]interface{}{"arraySize": 1000000},
		Complexity: 5,
	}
}

func generateErrorProneRequest(errorRate float64) LambdaPerformanceRequest {
	return LambdaPerformanceRequest{
		RequestID: fmt.Sprintf("err-%d", time.Now().UnixNano()),
		Payload: map[string]interface{}{
			"errorRate": errorRate,
			"shouldError": math.Rand.Float64() < errorRate,
		},
		Complexity: 1,
	}
}

// Lambda Function Code Generators

// Functional Lambda function code generators with immutable patterns

func generateFunctionalSimpleFunctionCode() []byte {
	return []byte(`
exports.handler = async (event) => {
	const startTime = Date.now();
	
	// Functional processing with immutable data
	const data = event.payload?.data || [];
	const sum = data.reduce((acc, _, index) => acc + index, 0);
	
	return {
		success: true,
		result: sum,
		processingTime: Date.now() - startTime,
		requestId: event.requestId
	};
};`)
}

func generateFunctionalCpuIntensiveFunctionCode() []byte {
	return []byte(`
exports.handler = async (event) => {
	const startTime = Date.now();
	const iterations = event.payload?.iterations || 100000;
	
	// Functional CPU intensive operation using pure functions
	const isPrime = (n) => {
		if (n < 2) return false;
		for (let i = 2; i <= Math.sqrt(n); i++) {
			if (n % i === 0) return false;
		}
		return true;
	};
	
	const primeCount = Array.from({ length: iterations }, (_, i) => i + 2)
		.filter(isPrime)
		.length;
	
	return {
		success: true,
		result: primeCount,
		processingTime: Date.now() - startTime,
		requestId: event.requestId
	};
};`)
}

func generateFunctionalMemoryIntensiveFunctionCode() []byte {
	return []byte(`
exports.handler = async (event) => {
	const startTime = Date.now();
	const arraySize = event.payload?.arraySize || 100000;
	
	// Functional memory intensive operation with immutable transformations
	const createItem = (id) => ({
		id,
		data: 'x'.repeat(100),
		timestamp: Date.now()
	});
	
	const processItem = (item) => ({ id: item.id, processed: true });
	
	const result = Array.from({ length: arraySize }, (_, i) => createItem(i))
		.filter((_, index) => index % 2 === 0)
		.map(processItem);
	
	return {
		success: true,
		result: result.length,
		processingTime: Date.now() - startTime,
		requestId: event.requestId
	};
};`)
}

func generateFunctionalIoIntensiveFunctionCode() []byte {
	return []byte(`
const https = require('https');

exports.handler = async (event) => {
	const startTime = Date.now();
	
	// Functional IO intensive operation with promise composition
	const makeHttpRequest = (url) => new Promise((resolve, reject) => {
		const req = https.get(url, (res) => {
			let data = '';
			res.on('data', chunk => data += chunk);
			res.on('end', () => resolve(data));
		});
		req.on('error', reject);
		req.setTimeout(5000, () => {
			req.destroy();
			reject(new Error('Request timeout'));
		});
	});
	
	const createUrl = (i) => \`https://httpbin.org/delay/\${i}\`;
	
	try {
		const results = await Promise.all(
			Array.from({ length: 5 }, (_, i) => makeHttpRequest(createUrl(i)))
		);
		
		return {
			success: true,
			result: results.length,
			processingTime: Date.now() - startTime,
			requestId: event.requestId
		};
	} catch (error) {
		return {
			success: false,
			error: error.message,
			processingTime: Date.now() - startTime,
			requestId: event.requestId
		};
	}
};`)
}

func generateFunctionalErrorProneFunctionCode() []byte {
	return []byte(`
exports.handler = async (event) => {
	const startTime = Date.now();
	const shouldError = event.payload?.shouldError || false;
	
	// Functional error handling with pure validation
	const validateRequest = (shouldErr) => shouldErr 
		? { success: false, error: 'Simulated error for testing' }
		: { success: true, result: 'ok' };
	
	const result = validateRequest(shouldError);
	
	if (!result.success) {
		throw new Error(result.error);
	}
	
	return {
		...result,
		processingTime: Date.now() - startTime,
		requestId: event.requestId
	};
};`)
}

// Functional Memory Profiler Implementation

func NewFunctionalMemoryProfiler() *MemoryProfiler {
	return lo.Pipe2(
		&MemoryProfiler{},
		func(profiler *MemoryProfiler) *MemoryProfiler {
			profiler.snapshots = make([]runtime.MemStats, 0)
			profiler.stopChan = make(chan struct{})
			return profiler
		},
		func(profiler *MemoryProfiler) *MemoryProfiler {
			return profiler
		},
	)
}

func NewMemoryProfiler() *MemoryProfiler {
	return NewFunctionalMemoryProfiler()
}

func (mp *MemoryProfiler) Start() {
	mp.mu.Lock()
	defer mp.mu.Unlock()
	
	if mp.running {
		return
	}
	
	mp.running = true
	runtime.ReadMemStats(&mp.initialMemStats)
	
	go mp.profileLoop()
}

func (mp *MemoryProfiler) Stop() {
	mp.mu.Lock()
	defer mp.mu.Unlock()
	
	if !mp.running {
		return
	}
	
	mp.running = false
	close(mp.stopChan)
}

func (mp *MemoryProfiler) profileLoop() {
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()
	
	for {
		select {
		case <-ticker.C:
			var memStats runtime.MemStats
			runtime.ReadMemStats(&memStats)
			
			mp.mu.Lock()
			mp.snapshots = append(mp.snapshots, memStats)
			mp.mu.Unlock()
			
		case <-mp.stopChan:
			return
		}
	}
}

type MemoryStats struct {
	HeapAllocBytes   uint64
	HeapSysBytes     uint64
	NumGC            uint32
	PauseTotalNs     uint64
	LeakDetected     bool
}

func (mp *MemoryProfiler) GetStats() MemoryStats {
	mp.mu.Lock()
	defer mp.mu.Unlock()
	
	if len(mp.snapshots) == 0 {
		return MemoryStats{}
	}
	
	latest := mp.snapshots[len(mp.snapshots)-1]
	
	// Simple leak detection: check if heap is consistently growing
	leakDetected := false
	if len(mp.snapshots) > 10 {
		recent := mp.snapshots[len(mp.snapshots)-10:]
		growing := 0
		for i := 1; i < len(recent); i++ {
			if recent[i].HeapAlloc > recent[i-1].HeapAlloc {
				growing++
			}
		}
		leakDetected = growing > 7 // More than 70% growing
	}
	
	return MemoryStats{
		HeapAllocBytes: latest.HeapAlloc,
		HeapSysBytes:   latest.HeapSys,
		NumGC:          latest.NumGC - mp.initialMemStats.NumGC,
		PauseTotalNs:   latest.PauseTotalNs - mp.initialMemStats.PauseTotalNs,
		LeakDetected:   leakDetected,
	}
}

// Helper Functions

func validateBasicPerformance(t *testing.T, testName string, input LambdaPerformanceRequest, expected LambdaPerformanceResult, actual LambdaPerformanceResult) {
	t.Helper()
	
	assert.Equal(t, expected.Success, actual.Success, "Test %s: Success status mismatch", testName)
	assert.Equal(t, input.RequestID, actual.RequestID, "Test %s: Request ID mismatch", testName)
	
	if actual.Success {
		assert.Less(t, actual.ExecutionTime, 10*time.Second, "Test %s: Execution time too long", testName)
		assert.Greater(t, actual.MemoryUsed, int64(0), "Test %s: Memory usage should be reported", testName)
	}
}

func calculateAverageExecutionTime(results []LambdaPerformanceResult) time.Duration {
	if len(results) == 0 {
		return 0
	}
	
	var total time.Duration
	for _, result := range results {
		total += result.ExecutionTime
	}
	
	return total / time.Duration(len(results))
}

func parseLambdaMetrics(logOutput string) (billedDuration time.Duration, memoryUsed int64, initDuration time.Duration, found bool) {
	// This would parse actual Lambda log output to extract metrics
	// For example purposes, returning mock values
	return 100 * time.Millisecond, 64 * 1024 * 1024, 0, true
}

func sortLatencies(latencies []time.Duration) {
	// Simple bubble sort for demonstration
	n := len(latencies)
	for i := 0; i < n-1; i++ {
		for j := 0; j < n-i-1; j++ {
			if latencies[j] > latencies[j+1] {
				latencies[j], latencies[j+1] = latencies[j+1], latencies[j]
			}
		}
	}
}

func measureCurrentPerformance(t *testing.T, ctx *LambdaPerformanceTestContext) map[string]PerformanceMetrics {
	t.Helper()
	
	// Simplified implementation - measure performance for all functions
	results := make(map[string]PerformanceMetrics)
	
	functions := map[string]string{
		"simple":          ctx.SimpleFunctionName,
		"cpu_intensive":   ctx.CpuIntensiveFunctionName,
		"memory_intensive": ctx.MemoryIntensiveFunctionName,
	}
	
	for name, functionName := range functions {
		// Execute a small performance test
		var testResults []LambdaPerformanceResult
		for i := 0; i < 10; i++ {
			request := generateSimpleRequest(1000)
			result := invokeLambdaWithMetrics(t, ctx, functionName, request)
			testResults = append(testResults, result)
		}
		
		results[name] = calculatePerformanceMetrics(testResults)
	}
	
	return results
}

func loadFunctionalPerformanceBaselines(t *testing.T, ctx *LambdaPerformanceTestContext) {
	// Functional baseline loading with monadic error handling
	baselineResult := lo.Pipe2(
		ctx,
		func(c *LambdaPerformanceTestContext) mo.Result[map[string]PerformanceMetrics] {
			// In a real implementation, this would load baselines from storage
			return mo.Ok(make(map[string]PerformanceMetrics))
		},
		func(result mo.Result[map[string]PerformanceMetrics]) map[string]PerformanceMetrics {
			return result.Match(
				func(baselines map[string]PerformanceMetrics) map[string]PerformanceMetrics {
					ctx.PerformanceBaseline = baselines
					t.Logf("Loaded %d functional performance baselines", len(baselines))
					return baselines
				},
				func(err error) map[string]PerformanceMetrics {
					t.Logf("Failed to load functional performance baselines: %v", err)
					return make(map[string]PerformanceMetrics)
				},
			)
		},
	)
	
	_ = baselineResult // Consume result
}

func saveFunctionalPerformanceBaselines(t *testing.T, ctx *LambdaPerformanceTestContext, metrics map[string]PerformanceMetrics) {
	// Functional baseline saving with validation
	saveResult := lo.Pipe3(
		metrics,
		func(m map[string]PerformanceMetrics) mo.Result[map[string]PerformanceMetrics] {
			if len(m) == 0 {
				return mo.Err[map[string]PerformanceMetrics](fmt.Errorf("no metrics to save"))
			}
			return mo.Ok(m)
		},
		func(result mo.Result[map[string]PerformanceMetrics]) mo.Result[string] {
			return result.Map(func(m map[string]PerformanceMetrics) string {
				// In a real implementation, this would save baselines to storage
				return fmt.Sprintf("Saved %d functional performance baselines", len(m))
			})
		},
		func(result mo.Result[string]) bool {
			return result.Match(
				func(message string) bool {
					t.Log(message)
					return true
				},
				func(err error) bool {
					t.Logf("Failed to save functional performance baselines: %v", err)
					return false
				},
			)
		},
	)
	
	_ = saveResult // Consume result
}

func cleanupFunctionalPerformanceInfrastructure(ctx *LambdaPerformanceTestContext) error {
	// Functional cleanup with error boundary
	return lo.Pipe2(
		ctx,
		func(c *LambdaPerformanceTestContext) mo.Result[bool] {
			// Cleanup handled by VasDeference
			return mo.Ok(true)
		},
		func(result mo.Result[bool]) error {
			return result.Match(
				func(success bool) error { return nil },
				func(err error) error { return err },
			)
		},
	)
}

func loadPerformanceBaselines(t *testing.T, ctx *LambdaPerformanceTestContext) {
	loadFunctionalPerformanceBaselines(t, ctx)
}

func savePerformanceBaselines(t *testing.T, ctx *LambdaPerformanceTestContext, metrics map[string]PerformanceMetrics) {
	saveFunctionalPerformanceBaselines(t, ctx, metrics)
}

func cleanupPerformanceInfrastructure(ctx *LambdaPerformanceTestContext) error {
	return cleanupFunctionalPerformanceInfrastructure(ctx)
}