// Package performance - Comprehensive benchmarks for the VasDeference framework
// This file contains Go benchmark tests to measure actual performance characteristics
package performance

import (
	"runtime"
	"sync"
	"testing"
	"time"
	"vasdeference"
)

// BenchmarkMemoryPool tests memory pool performance vs standard allocation
func BenchmarkMemoryPool(b *testing.B) {
	pool := NewMemoryPool[[]byte](func() []byte {
		return make([]byte, 1024)
	})
	
	b.Run("WithPool", func(b *testing.B) {
		b.ResetTimer()
		b.ReportAllocs()
		
		for i := 0; i < b.N; i++ {
			data := pool.Get()
			// Simulate some work
			_ = append(data[:0], []byte("benchmark data")...)
			pool.Put(data)
		}
	})
	
	b.Run("WithoutPool", func(b *testing.B) {
		b.ResetTimer()
		b.ReportAllocs()
		
		for i := 0; i < b.N; i++ {
			data := make([]byte, 1024)
			// Simulate same work
			_ = append(data[:0], []byte("benchmark data")...)
		}
	})
}

// BenchmarkPerformanceMonitor tests monitoring overhead
func BenchmarkPerformanceMonitor(b *testing.B) {
	monitor := NewPerformanceMonitor()
	monitor.Start()
	defer monitor.Stop()
	
	b.Run("RecordLatency", func(b *testing.B) {
		b.ResetTimer()
		
		for i := 0; i < b.N; i++ {
			monitor.RecordLatency("benchmark_operation", 5*time.Millisecond)
		}
	})
	
	b.Run("RecordError", func(b *testing.B) {
		b.ResetTimer()
		
		for i := 0; i < b.N; i++ {
			monitor.RecordError("benchmark_operation")
		}
	})
	
	b.Run("GetMetrics", func(b *testing.B) {
		// Warm up with some data
		for i := 0; i < 1000; i++ {
			monitor.RecordLatency("benchmark_operation", time.Duration(i)*time.Microsecond)
		}
		
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = monitor.GetMetrics()
		}
	})
}

// BenchmarkConcurrentMonitoring tests concurrent monitoring performance
func BenchmarkConcurrentMonitoring(b *testing.B) {
	monitor := NewPerformanceMonitor()
	monitor.Start()
	defer monitor.Stop()
	
	b.Run("ConcurrentRecordLatency", func(b *testing.B) {
		b.ResetTimer()
		
		var wg sync.WaitGroup
		for i := 0; i < runtime.GOMAXPROCS(0); i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for j := 0; j < b.N/runtime.GOMAXPROCS(0); j++ {
					monitor.RecordLatency("concurrent_benchmark", 10*time.Microsecond)
				}
			}()
		}
		wg.Wait()
	})
}

// BenchmarkTableTestFramework benchmarks the table testing framework performance
func BenchmarkTableTestFramework(b *testing.B) {
	mockT := &benchTestingT{}
	
	// Simple operation benchmark
	b.Run("SimpleOperation", func(b *testing.B) {
		b.ResetTimer()
		
		for i := 0; i < b.N; i++ {
			runner := vasdeference.TableTest[int, int](mockT, "benchmark_simple")
			for j := 0; j < 10; j++ {
				runner.Case("case", j, j*j)
			}
			
			runner.Run(func(x int) int {
				return x * x
			}, func(name string, input, expected, actual int) {
				// Minimal validation
			})
		}
	})
	
	// Parallel operation benchmark
	b.Run("ParallelOperation", func(b *testing.B) {
		b.ResetTimer()
		
		for i := 0; i < b.N; i++ {
			runner := vasdeference.TableTest[int, int](mockT, "benchmark_parallel")
			for j := 0; j < 10; j++ {
				runner.Case("case", j, j*j)
			}
			
			runner.Parallel().WithMaxWorkers(4).Run(func(x int) int {
				return x * x
			}, func(name string, input, expected, actual int) {
				// Minimal validation
			})
		}
	})
}

// BenchmarkRegressionDetection tests regression detection performance
func BenchmarkRegressionDetection(b *testing.B) {
	detector := NewRegressionDetector()
	
	baseline := PerformanceMetrics{
		TestName:       "benchmark_regression",
		AverageLatency: 100 * time.Millisecond,
		P95Latency:     200 * time.Millisecond,
		Throughput:     1000.0,
		ErrorRate:      0.01,
		MemoryUsed:     50 * 1024 * 1024,
	}
	
	current := PerformanceMetrics{
		TestName:       "benchmark_regression",
		AverageLatency: 120 * time.Millisecond,
		P95Latency:     240 * time.Millisecond,
		Throughput:     850.0,
		ErrorRate:      0.015,
		MemoryUsed:     55 * 1024 * 1024,
	}
	
	detector.SetBaseline("benchmark_regression", baseline)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = detector.CheckRegression("benchmark_regression", current)
	}
}

// BenchmarkLoadTestExecution tests load test framework performance
func BenchmarkLoadTestExecution(b *testing.B) {
	if testing.Short() {
		b.Skip("Skipping load test benchmark in short mode")
	}
	
	config := LoadTestConfig{
		TestName:       "benchmark_load_test",
		Duration:       1 * time.Second,
		RequestsPerSec: 100,
		Concurrency:    10,
		TestFunction: func() error {
			// Simulate minimal work
			time.Sleep(1 * time.Microsecond)
			return nil
		},
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ExecuteLoadTest(config)
	}
}

// benchTestingT minimal implementation for benchmarking
type benchTestingT struct {
	errored bool
}

func (b *benchTestingT) Helper()                                    {}
func (b *benchTestingT) Errorf(format string, args ...interface{}) { b.errored = true }
func (b *benchTestingT) Error(args ...interface{})                 { b.errored = true }
func (b *benchTestingT) Fail()                                     {}
func (b *benchTestingT) FailNow()                                  {}
func (b *benchTestingT) Fatal(args ...interface{})                 {}
func (b *benchTestingT) Fatalf(format string, args ...interface{}) {}
func (b *benchTestingT) Logf(format string, args ...interface{})   {}
func (b *benchTestingT) Name() string                              { return "BenchTest" }