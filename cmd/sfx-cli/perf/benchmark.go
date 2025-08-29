package perf

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"runtime"
	"sort"
	"sync"
	"time"

	"github.com/fatih/color"
)

// BenchmarkSuite represents a collection of performance benchmarks
type BenchmarkSuite struct {
	Name        string
	Service     string
	Duration    time.Duration
	Concurrency int
	Results     []*BenchmarkResult
	
	green  *color.Color
	red    *color.Color
	yellow *color.Color
	blue   *color.Color
	cyan   *color.Color
}

// BenchmarkResult contains the results of a performance benchmark
type BenchmarkResult struct {
	Name           string            `json:"name"`
	Service        string            `json:"service"`
	Operation      string            `json:"operation"`
	TotalRequests  int64             `json:"total_requests"`
	SuccessCount   int64             `json:"success_count"`
	ErrorCount     int64             `json:"error_count"`
	Duration       time.Duration     `json:"duration"`
	RequestsPerSec float64           `json:"requests_per_sec"`
	AvgLatency     time.Duration     `json:"avg_latency"`
	MinLatency     time.Duration     `json:"min_latency"`
	MaxLatency     time.Duration     `json:"max_latency"`
	P50Latency     time.Duration     `json:"p50_latency"`
	P90Latency     time.Duration     `json:"p90_latency"`
	P95Latency     time.Duration     `json:"p95_latency"`
	P99Latency     time.Duration     `json:"p99_latency"`
	Errors         []string          `json:"errors"`
	Metadata       map[string]string `json:"metadata"`
	Timestamp      time.Time         `json:"timestamp"`
}

// LatencyData holds individual request latency measurements
type LatencyData struct {
	Latencies []time.Duration
	mutex     sync.RWMutex
}

// BenchmarkConfig configures benchmark execution
type BenchmarkConfig struct {
	Service     string
	Operation   string
	Duration    time.Duration
	Concurrency int
	WarmupTime  time.Duration
	Payload     interface{}
	Headers     map[string]string
	Region      string
	Endpoint    string
}

// NewBenchmarkSuite creates a new benchmark suite
func NewBenchmarkSuite(name, service string, duration time.Duration, concurrency int) *BenchmarkSuite {
	return &BenchmarkSuite{
		Name:        name,
		Service:     service,
		Duration:    duration,
		Concurrency: concurrency,
		Results:     make([]*BenchmarkResult, 0),
		green:       color.New(color.FgGreen),
		red:         color.New(color.FgRed),
		yellow:      color.New(color.FgYellow),
		blue:        color.New(color.FgBlue),
		cyan:        color.New(color.FgCyan),
	}
}

// Run executes the benchmark suite
func (bs *BenchmarkSuite) Run(ctx context.Context) error {
	bs.printBanner()
	
	switch bs.Service {
	case "lambda":
		return bs.runLambdaBenchmarks(ctx)
	case "dynamodb":
		return bs.runDynamoDBBenchmarks(ctx)
	case "eventbridge":
		return bs.runEventBridgeBenchmarks(ctx)
	case "stepfunctions":
		return bs.runStepFunctionsBenchmarks(ctx)
	default:
		return bs.runGenericBenchmarks(ctx)
	}
}

// printBanner displays the benchmark suite banner
func (bs *BenchmarkSuite) printBanner() {
	bs.cyan.Println(`
╔══════════════════════════════════════════════════════════════════════════════╗
║                          SFX Performance Benchmark                          ║
║                       Serverless Framework eXtended                         ║
╚══════════════════════════════════════════════════════════════════════════════╝
`)
	
	fmt.Printf("🎯 Suite: %s\n", bs.Name)
	fmt.Printf("📊 Service: %s\n", bs.Service)
	fmt.Printf("⏱️  Duration: %v\n", bs.Duration)
	fmt.Printf("🚀 Concurrency: %d\n", bs.Concurrency)
	fmt.Printf("💻 CPUs: %d\n", runtime.NumCPU())
	fmt.Println()
}

// runLambdaBenchmarks executes Lambda-specific benchmarks
func (bs *BenchmarkSuite) runLambdaBenchmarks(ctx context.Context) error {
	bs.blue.Println("🔧 Running Lambda Performance Benchmarks...")
	
	benchmarks := []BenchmarkConfig{
		{
			Service:     "lambda",
			Operation:   "InvokeFunction",
			Duration:    bs.Duration,
			Concurrency: bs.Concurrency,
			WarmupTime:  10 * time.Second,
			Payload:     map[string]interface{}{"test": "payload"},
		},
		{
			Service:     "lambda",
			Operation:   "InvokeFunctionAsync",
			Duration:    bs.Duration,
			Concurrency: bs.Concurrency,
			WarmupTime:  5 * time.Second,
			Payload:     map[string]interface{}{"async": "payload"},
		},
	}
	
	for _, config := range benchmarks {
		result, err := bs.runSingleBenchmark(ctx, config)
		if err != nil {
			bs.red.Printf("❌ Benchmark %s failed: %v\n", config.Operation, err)
			continue
		}
		bs.Results = append(bs.Results, result)
	}
	
	return nil
}

// runDynamoDBBenchmarks executes DynamoDB-specific benchmarks
func (bs *BenchmarkSuite) runDynamoDBBenchmarks(ctx context.Context) error {
	bs.blue.Println("🗄️  Running DynamoDB Performance Benchmarks...")
	
	benchmarks := []BenchmarkConfig{
		{
			Service:     "dynamodb",
			Operation:   "PutItem",
			Duration:    bs.Duration,
			Concurrency: bs.Concurrency,
			WarmupTime:  5 * time.Second,
		},
		{
			Service:     "dynamodb",
			Operation:   "GetItem",
			Duration:    bs.Duration,
			Concurrency: bs.Concurrency,
			WarmupTime:  5 * time.Second,
		},
		{
			Service:     "dynamodb",
			Operation:   "Query",
			Duration:    bs.Duration,
			Concurrency: bs.Concurrency / 2, // Reduce concurrency for queries
			WarmupTime:  5 * time.Second,
		},
	}
	
	for _, config := range benchmarks {
		result, err := bs.runSingleBenchmark(ctx, config)
		if err != nil {
			bs.red.Printf("❌ Benchmark %s failed: %v\n", config.Operation, err)
			continue
		}
		bs.Results = append(bs.Results, result)
	}
	
	return nil
}

// runEventBridgeBenchmarks executes EventBridge-specific benchmarks
func (bs *BenchmarkSuite) runEventBridgeBenchmarks(ctx context.Context) error {
	bs.blue.Println("📡 Running EventBridge Performance Benchmarks...")
	
	benchmarks := []BenchmarkConfig{
		{
			Service:     "eventbridge",
			Operation:   "PutEvents",
			Duration:    bs.Duration,
			Concurrency: bs.Concurrency,
			WarmupTime:  5 * time.Second,
		},
	}
	
	for _, config := range benchmarks {
		result, err := bs.runSingleBenchmark(ctx, config)
		if err != nil {
			bs.red.Printf("❌ Benchmark %s failed: %v\n", config.Operation, err)
			continue
		}
		bs.Results = append(bs.Results, result)
	}
	
	return nil
}

// runStepFunctionsBenchmarks executes Step Functions benchmarks
func (bs *BenchmarkSuite) runStepFunctionsBenchmarks(ctx context.Context) error {
	bs.blue.Println("🔀 Running Step Functions Performance Benchmarks...")
	
	benchmarks := []BenchmarkConfig{
		{
			Service:     "stepfunctions",
			Operation:   "StartExecution",
			Duration:    bs.Duration,
			Concurrency: bs.Concurrency / 4, // Reduce concurrency for state machines
			WarmupTime:  10 * time.Second,
		},
	}
	
	for _, config := range benchmarks {
		result, err := bs.runSingleBenchmark(ctx, config)
		if err != nil {
			bs.red.Printf("❌ Benchmark %s failed: %v\n", config.Operation, err)
			continue
		}
		bs.Results = append(bs.Results, result)
	}
	
	return nil
}

// runGenericBenchmarks executes generic performance tests
func (bs *BenchmarkSuite) runGenericBenchmarks(ctx context.Context) error {
	bs.blue.Println("🔧 Running Generic Performance Benchmarks...")
	
	config := BenchmarkConfig{
		Service:     bs.Service,
		Operation:   "Generic",
		Duration:    bs.Duration,
		Concurrency: bs.Concurrency,
		WarmupTime:  5 * time.Second,
	}
	
	result, err := bs.runSingleBenchmark(ctx, config)
	if err != nil {
		return fmt.Errorf("generic benchmark failed: %w", err)
	}
	
	bs.Results = append(bs.Results, result)
	return nil
}

// runSingleBenchmark executes a single benchmark configuration
func (bs *BenchmarkSuite) runSingleBenchmark(ctx context.Context, config BenchmarkConfig) (*BenchmarkResult, error) {
	bs.cyan.Printf("🧪 Running %s benchmark...\n", config.Operation)
	
	// Initialize result
	result := &BenchmarkResult{
		Name:      fmt.Sprintf("%s_%s", config.Service, config.Operation),
		Service:   config.Service,
		Operation: config.Operation,
		Errors:    make([]string, 0),
		Metadata:  make(map[string]string),
		Timestamp: time.Now(),
	}
	
	// Initialize latency tracking
	latencyData := &LatencyData{
		Latencies: make([]time.Duration, 0),
	}
	
	// Warmup phase
	if config.WarmupTime > 0 {
		bs.yellow.Printf("🔥 Warming up for %v...\n", config.WarmupTime)
		warmupCtx, cancel := context.WithTimeout(ctx, config.WarmupTime)
		bs.runWarmup(warmupCtx, config)
		cancel()
	}
	
	// Main benchmark execution
	bs.green.Printf("🏁 Starting benchmark (%v, %d workers)...\n", config.Duration, config.Concurrency)
	
	benchmarkCtx, cancel := context.WithTimeout(ctx, config.Duration)
	defer cancel()
	
	var wg sync.WaitGroup
	var requestCount, successCount, errorCount int64
	var errorList []string
	var errorMutex sync.Mutex
	
	startTime := time.Now()
	
	// Launch concurrent workers
	for i := 0; i < config.Concurrency; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			
			for {
				select {
				case <-benchmarkCtx.Done():
					return
				default:
					// Execute single operation
					opStart := time.Now()
					err := bs.executeOperation(ctx, config)
					latency := time.Since(opStart)
					
					// Record latency
					latencyData.mutex.Lock()
					latencyData.Latencies = append(latencyData.Latencies, latency)
					latencyData.mutex.Unlock()
					
					// Update counters
					if err != nil {
						atomic.AddInt64(&errorCount, 1)
						errorMutex.Lock()
						errorList = append(errorList, err.Error())
						errorMutex.Unlock()
					} else {
						atomic.AddInt64(&successCount, 1)
					}
					
					atomic.AddInt64(&requestCount, 1)
				}
			}
		}(i)
	}
	
	// Wait for completion
	wg.Wait()
	actualDuration := time.Since(startTime)
	
	// Calculate statistics
	result.TotalRequests = requestCount
	result.SuccessCount = successCount
	result.ErrorCount = errorCount
	result.Duration = actualDuration
	result.RequestsPerSec = float64(requestCount) / actualDuration.Seconds()
	result.Errors = errorList
	
	// Calculate latency statistics
	bs.calculateLatencyStats(result, latencyData)
	
	// Print results
	bs.printBenchmarkResult(result)
	
	return result, nil
}

// executeOperation executes a single operation based on service type
func (bs *BenchmarkSuite) executeOperation(ctx context.Context, config BenchmarkConfig) error {
	// Simulate operation execution with random latency
	switch config.Service {
	case "lambda":
		return bs.simulateLambdaOperation(ctx, config)
	case "dynamodb":
		return bs.simulateDynamoDBOperation(ctx, config)
	case "eventbridge":
		return bs.simulateEventBridgeOperation(ctx, config)
	case "stepfunctions":
		return bs.simulateStepFunctionsOperation(ctx, config)
	default:
		return bs.simulateGenericOperation(ctx, config)
	}
}

// simulateLambdaOperation simulates Lambda function invocation
func (bs *BenchmarkSuite) simulateLambdaOperation(ctx context.Context, config BenchmarkConfig) error {
	// Simulate variable Lambda execution time
	baseLatency := time.Duration(50+rand.Intn(200)) * time.Millisecond
	time.Sleep(baseLatency)
	
	// Simulate occasional errors (5% error rate)
	if rand.Float64() < 0.05 {
		return fmt.Errorf("simulated lambda error")
	}
	
	return nil
}

// simulateDynamoDBOperation simulates DynamoDB operations
func (bs *BenchmarkSuite) simulateDynamoDBOperation(ctx context.Context, config BenchmarkConfig) error {
	var baseLatency time.Duration
	
	switch config.Operation {
	case "PutItem":
		baseLatency = time.Duration(10+rand.Intn(20)) * time.Millisecond
	case "GetItem":
		baseLatency = time.Duration(5+rand.Intn(15)) * time.Millisecond
	case "Query":
		baseLatency = time.Duration(20+rand.Intn(50)) * time.Millisecond
	default:
		baseLatency = time.Duration(10+rand.Intn(30)) * time.Millisecond
	}
	
	time.Sleep(baseLatency)
	
	// Simulate occasional throttling (2% error rate)
	if rand.Float64() < 0.02 {
		return fmt.Errorf("simulated throttling error")
	}
	
	return nil
}

// simulateEventBridgeOperation simulates EventBridge operations
func (bs *BenchmarkSuite) simulateEventBridgeOperation(ctx context.Context, config BenchmarkConfig) error {
	// EventBridge is generally fast
	baseLatency := time.Duration(5+rand.Intn(10)) * time.Millisecond
	time.Sleep(baseLatency)
	
	// Very low error rate for EventBridge
	if rand.Float64() < 0.01 {
		return fmt.Errorf("simulated eventbridge error")
	}
	
	return nil
}

// simulateStepFunctionsOperation simulates Step Functions operations
func (bs *BenchmarkSuite) simulateStepFunctionsOperation(ctx context.Context, config BenchmarkConfig) error {
	// Step Functions have higher latency
	baseLatency := time.Duration(100+rand.Intn(300)) * time.Millisecond
	time.Sleep(baseLatency)
	
	// Moderate error rate
	if rand.Float64() < 0.03 {
		return fmt.Errorf("simulated step functions error")
	}
	
	return nil
}

// simulateGenericOperation simulates generic operations
func (bs *BenchmarkSuite) simulateGenericOperation(ctx context.Context, config BenchmarkConfig) error {
	baseLatency := time.Duration(25+rand.Intn(75)) * time.Millisecond
	time.Sleep(baseLatency)
	
	if rand.Float64() < 0.03 {
		return fmt.Errorf("simulated generic error")
	}
	
	return nil
}

// runWarmup executes warmup phase
func (bs *BenchmarkSuite) runWarmup(ctx context.Context, config BenchmarkConfig) {
	var wg sync.WaitGroup
	
	for i := 0; i < config.Concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			
			for {
				select {
				case <-ctx.Done():
					return
				default:
					bs.executeOperation(ctx, config)
					time.Sleep(100 * time.Millisecond)
				}
			}
		}()
	}
	
	wg.Wait()
}

// calculateLatencyStats calculates latency percentiles and statistics
func (bs *BenchmarkSuite) calculateLatencyStats(result *BenchmarkResult, data *LatencyData) {
	data.mutex.RLock()
	defer data.mutex.RUnlock()
	
	if len(data.Latencies) == 0 {
		return
	}
	
	// Sort latencies for percentile calculations
	sort.Slice(data.Latencies, func(i, j int) bool {
		return data.Latencies[i] < data.Latencies[j]
	})
	
	n := len(data.Latencies)
	
	// Calculate basic stats
	result.MinLatency = data.Latencies[0]
	result.MaxLatency = data.Latencies[n-1]
	
	// Calculate average
	var total time.Duration
	for _, latency := range data.Latencies {
		total += latency
	}
	result.AvgLatency = total / time.Duration(n)
	
	// Calculate percentiles
	result.P50Latency = data.Latencies[n*50/100]
	result.P90Latency = data.Latencies[n*90/100]
	result.P95Latency = data.Latencies[n*95/100]
	result.P99Latency = data.Latencies[n*99/100]
}

// printBenchmarkResult displays the results of a single benchmark
func (bs *BenchmarkSuite) printBenchmarkResult(result *BenchmarkResult) {
	fmt.Println()
	bs.green.Printf("✅ %s Results:\n", result.Name)
	fmt.Println(strings.Repeat("─", 60))
	
	fmt.Printf("📊 Total Requests:    %d\n", result.TotalRequests)
	fmt.Printf("✅ Successful:        %d\n", result.SuccessCount)
	fmt.Printf("❌ Errors:           %d\n", result.ErrorCount)
	fmt.Printf("⚡ Requests/sec:      %.2f\n", result.RequestsPerSec)
	fmt.Printf("⏱️  Duration:          %v\n", result.Duration.Truncate(time.Millisecond))
	
	fmt.Println()
	fmt.Println("📈 Latency Statistics:")
	fmt.Printf("   Average:    %v\n", result.AvgLatency.Truncate(time.Microsecond))
	fmt.Printf("   Minimum:    %v\n", result.MinLatency.Truncate(time.Microsecond))
	fmt.Printf("   Maximum:    %v\n", result.MaxLatency.Truncate(time.Microsecond))
	fmt.Printf("   P50:        %v\n", result.P50Latency.Truncate(time.Microsecond))
	fmt.Printf("   P90:        %v\n", result.P90Latency.Truncate(time.Microsecond))
	fmt.Printf("   P95:        %v\n", result.P95Latency.Truncate(time.Microsecond))
	fmt.Printf("   P99:        %v\n", result.P99Latency.Truncate(time.Microsecond))
	
	if result.ErrorCount > 0 {
		fmt.Println()
		bs.red.Println("❌ Error Summary:")
		errorCounts := make(map[string]int)
		for _, err := range result.Errors {
			errorCounts[err]++
		}
		
		for errMsg, count := range errorCounts {
			fmt.Printf("   %s: %d\n", errMsg, count)
		}
	}
	
	fmt.Println(strings.Repeat("─", 60))
}

// PrintSummary displays overall benchmark suite results
func (bs *BenchmarkSuite) PrintSummary() {
	fmt.Println()
	bs.cyan.Println("📊 Benchmark Suite Summary")
	fmt.Println(strings.Repeat("═", 80))
	
	if len(bs.Results) == 0 {
		bs.yellow.Println("⚠️  No benchmark results available")
		return
	}
	
	var totalRequests, totalSuccess, totalErrors int64
	var totalDuration time.Duration
	
	for _, result := range bs.Results {
		totalRequests += result.TotalRequests
		totalSuccess += result.SuccessCount
		totalErrors += result.ErrorCount
		totalDuration += result.Duration
	}
	
	fmt.Printf("🎯 Suite:             %s\n", bs.Name)
	fmt.Printf("📊 Service:           %s\n", bs.Service)
	fmt.Printf("🧪 Benchmarks:        %d\n", len(bs.Results))
	fmt.Printf("📊 Total Requests:    %d\n", totalRequests)
	fmt.Printf("✅ Success Rate:      %.2f%%\n", float64(totalSuccess)/float64(totalRequests)*100)
	
	if totalErrors > 0 {
		fmt.Printf("❌ Error Rate:        %.2f%%\n", float64(totalErrors)/float64(totalRequests)*100)
	}
	
	fmt.Println()
	
	// Show top performing benchmarks
	bs.showTopPerformers()
	
	fmt.Println(strings.Repeat("═", 80))
}

// showTopPerformers shows the best and worst performing benchmarks
func (bs *BenchmarkSuite) showTopPerformers() {
	if len(bs.Results) == 0 {
		return
	}
	
	// Sort by requests per second
	sortedResults := make([]*BenchmarkResult, len(bs.Results))
	copy(sortedResults, bs.Results)
	
	sort.Slice(sortedResults, func(i, j int) bool {
		return sortedResults[i].RequestsPerSec > sortedResults[j].RequestsPerSec
	})
	
	bs.green.Println("🏆 Top Performers:")
	for i, result := range sortedResults {
		if i >= 3 {
			break // Show top 3
		}
		
		status := "🥇"
		if i == 1 {
			status = "🥈"
		} else if i == 2 {
			status = "🥉"
		}
		
		fmt.Printf("   %s %s: %.2f req/s (avg: %v)\n", 
			status, result.Name, result.RequestsPerSec, result.AvgLatency.Truncate(time.Microsecond))
	}
}

// SaveResults saves benchmark results to a JSON file
func (bs *BenchmarkSuite) SaveResults(filename string) error {
	data := map[string]interface{}{
		"suite":       bs.Name,
		"service":     bs.Service,
		"duration":    bs.Duration,
		"concurrency": bs.Concurrency,
		"results":     bs.Results,
		"timestamp":   time.Now(),
		"system": map[string]interface{}{
			"cpus":    runtime.NumCPU(),
			"goarch":  runtime.GOARCH,
			"goos":    runtime.GOOS,
			"version": runtime.Version(),
		},
	}
	
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create results file: %w", err)
	}
	defer file.Close()
	
	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	
	if err := encoder.Encode(data); err != nil {
		return fmt.Errorf("failed to encode results: %w", err)
	}
	
	bs.green.Printf("💾 Results saved to: %s\n", filename)
	return nil
}

// Import missing packages for completeness
import (
	"math/rand"
	"strings"
	"sync/atomic"
	"time"
)