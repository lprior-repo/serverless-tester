# Performance Testing Examples

Advanced performance testing patterns using the enhanced VasDeference framework with benchmarking, load testing, and memory optimization.

## Overview

These examples demonstrate production-ready performance testing for serverless applications:

- **Benchmarking Patterns**: Precise performance measurement with statistical analysis
- **Load Testing**: High-concurrency testing with parallel execution control
- **Memory Profiling**: Memory usage optimization and leak detection
- **Latency Testing**: Response time validation across different scenarios
- **Throughput Testing**: Maximum capacity determination
- **Resource Utilization**: CPU, memory, and network usage analysis

## Examples

### 1. Lambda Function Performance Testing
**File**: `lambda_performance_test.go`

Comprehensive Lambda performance validation:
- Cold start vs warm start benchmarking
- Memory allocation optimization
- Concurrent execution performance
- Timeout and error rate analysis
- Memory leak detection

### 2. DynamoDB Performance Testing
**File**: `dynamodb_performance_test.go`

Database performance optimization:
- Read/write throughput testing
- Batch operation performance
- Query vs scan performance analysis
- Index performance validation
- Connection pooling optimization

### 3. Step Functions Performance Testing
**File**: `stepfunctions_performance_test.go`

Workflow performance optimization:
- Parallel execution benchmarking
- State transition performance
- Error handling overhead
- Long-running workflow testing
- Batch operation efficiency

### 4. EventBridge Performance Testing
**File**: `eventbridge_performance_test.go`

Event processing performance:
- Event publishing throughput
- Rule processing performance
- Cross-region event latency
- Event filtering performance
- Dead letter queue analysis

## Key Features

### Advanced Benchmarking
```go
// Statistical benchmark analysis
benchmark := vasdeference.TableTest[TestInput, TestOutput](t, "Performance Benchmark").
    Case("baseline", input, expected).
    Benchmark(testFunction, 10000)

// Performance assertions with statistical significance
assert.Less(t, benchmark.AverageTime, 100*time.Millisecond)
assert.Less(t, benchmark.P95Time, 200*time.Millisecond)
assert.Less(t, benchmark.P99Time, 500*time.Millisecond)
```

### Concurrent Load Testing
```go
// High-concurrency load testing
vasdeference.TableTest[LoadTestRequest, LoadTestResult](t, "Load Test").
    Parallel().
    WithMaxWorkers(100).
    Repeat(1000).
    Timeout(300 * time.Second).
    Run(performLoadTest, validateLoadTestResult)
```

### Memory Profiling
```go
// Memory usage tracking
profiler := vasdeference.StartMemoryProfiler(t)
defer profiler.Stop()

// Execute operations
results := executeOperations(input)

// Analyze memory usage
memStats := profiler.GetStats()
assert.Less(t, memStats.HeapAllocBytes, 100*1024*1024) // 100MB limit
assert.Zero(t, memStats.LeakDetected)
```

### Performance Regression Testing
```go
// Baseline comparison testing
baseline := loadPerformanceBaseline(t, "v1.0.0")
current := measureCurrentPerformance(t)

// Regression detection
assert.InDelta(t, baseline.AverageLatency, current.AverageLatency, 0.10) // 10% tolerance
assert.LessOrEqual(t, current.ErrorRate, baseline.ErrorRate*1.05) // 5% error rate tolerance
```

## Running Performance Tests

### Prerequisites
```bash
# Set performance test configuration
export PERFORMANCE_TEST_DURATION=300s
export PERFORMANCE_TEST_CONCURRENCY=50
export PERFORMANCE_TEST_BASELINE_VERSION=v1.0.0

# Ensure adequate AWS service limits
export AWS_LAMBDA_CONCURRENT_EXECUTIONS=1000
export AWS_DYNAMODB_MAX_RCU=40000
export AWS_DYNAMODB_MAX_WCU=40000
```

### Execute Tests
```bash
# Run all performance tests
go test -v -timeout=600s ./examples/performance/

# Run benchmarks with CPU profiling
go test -v -bench=. -cpuprofile=cpu.prof ./examples/performance/

# Run memory profiling
go test -v -bench=. -memprofile=mem.prof ./examples/performance/

# Run load tests
LOAD_TEST=true go test -v -timeout=1800s ./examples/performance/

# Generate performance reports
go test -v -bench=. -benchmem -count=5 ./examples/performance/ > performance_report.txt
```

### Continuous Performance Monitoring
```bash
# Performance regression detection
go test -v -bench=. -benchmem \
    -test.benchtime=30s \
    -test.count=3 \
    ./examples/performance/ \
    | tee current_performance.txt

# Compare with baseline
benchcmp baseline_performance.txt current_performance.txt
```

## Performance Metrics

### Lambda Function Metrics
- **Cold Start Time**: Time from invocation to first response
- **Warm Start Time**: Time for subsequent invocations
- **Memory Usage**: Peak and average memory consumption
- **Duration**: Function execution time
- **Error Rate**: Percentage of failed invocations
- **Throttle Rate**: Percentage of throttled invocations

### DynamoDB Metrics
- **Read Latency**: Average read operation time
- **Write Latency**: Average write operation time
- **Throughput**: Operations per second
- **Throttle Rate**: Percentage of throttled requests
- **Connection Pool Efficiency**: Connection reuse statistics

### Step Functions Metrics
- **Execution Duration**: Total workflow time
- **State Transition Time**: Time between states
- **Parallel Branch Performance**: Concurrent execution efficiency
- **Error Recovery Time**: Time to handle and recover from errors

### EventBridge Metrics
- **Event Publishing Latency**: Time to publish events
- **Rule Processing Time**: Time to evaluate rules
- **Cross-Region Latency**: Inter-region event delivery time
- **Throughput**: Events processed per second

## Optimization Patterns

### Memory Optimization
```go
// Pool-based memory management
var requestPool = sync.Pool{
    New: func() interface{} {
        return &Request{}
    },
}

func processRequest(input RequestData) Result {
    req := requestPool.Get().(*Request)
    defer requestPool.Put(req)
    
    req.Reset()
    req.LoadData(input)
    return req.Process()
}
```

### Connection Pooling
```go
// Optimized client configuration
client := dynamodb.NewClient(config, func(o *dynamodb.Options) {
    o.HTTPClient = &http.Client{
        Transport: &http.Transport{
            MaxIdleConns:        100,
            MaxIdleConnsPerHost: 20,
            IdleConnTimeout:     90 * time.Second,
        },
        Timeout: 30 * time.Second,
    }
})
```

### Batch Operation Optimization
```go
// Efficient batch processing
results := dynamodb.BatchWriteItems(ctx, requests, dynamodb.BatchOptions{
    MaxConcurrency: 25,
    RetryPolicy: dynamodb.ExponentialBackoff{
        MaxRetries: 3,
        BaseDelay:  100 * time.Millisecond,
    },
})
```

## Alerting and Monitoring

### Performance Thresholds
```go
// Define performance SLA thresholds
type PerformanceSLA struct {
    MaxAverageLatency time.Duration // 100ms
    MaxP95Latency     time.Duration // 500ms 
    MaxP99Latency     time.Duration // 1000ms
    MaxErrorRate      float64       // 0.01 (1%)
    MinThroughput     float64       // 1000 RPS
}

func validatePerformanceSLA(t *testing.T, metrics PerformanceMetrics, sla PerformanceSLA) {
    assert.LessOrEqual(t, metrics.AverageLatency, sla.MaxAverageLatency)
    assert.LessOrEqual(t, metrics.P95Latency, sla.MaxP95Latency)
    assert.LessOrEqual(t, metrics.P99Latency, sla.MaxP99Latency)
    assert.LessOrEqual(t, metrics.ErrorRate, sla.MaxErrorRate)
    assert.GreaterOrEqual(t, metrics.Throughput, sla.MinThroughput)
}
```

### CloudWatch Integration
```go
// Publish custom metrics to CloudWatch
err := cloudwatch.PutMetric(ctx, cloudwatch.MetricData{
    Namespace: "Performance/Testing",
    MetricName: "AverageLatency",
    Value:     benchmark.AverageTime.Seconds(),
    Unit:      "Seconds",
    Dimensions: map[string]string{
        "TestSuite": "lambda_performance",
        "Version":   "v1.0.0",
    },
})
```

## Capacity Planning

### Load Testing Scenarios
1. **Normal Load**: Expected production traffic
2. **Peak Load**: 2x normal load for peak periods
3. **Stress Load**: 5x normal load for capacity limits
4. **Spike Load**: 10x normal load for traffic spikes
5. **Endurance Load**: Normal load for extended periods

### Resource Scaling Validation
```go
func TestAutoScalingPerformance(t *testing.T) {
    // Test scaling from baseline to peak load
    baseline := executeLoadTest(t, 100) // 100 RPS
    peak := executeLoadTest(t, 1000)    // 1000 RPS
    
    // Validate scaling efficiency
    assert.Less(t, peak.ScalingTime, 60*time.Second)
    assert.Less(t, peak.AverageLatency, baseline.AverageLatency*1.5)
}
```

---

**Next**: Explore specific performance testing patterns for your serverless architecture.