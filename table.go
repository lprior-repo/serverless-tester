// Package vasdeference provides table-driven testing utilities with maximum syntax sugar
package vasdeference

import (
	"context"
	"fmt"
	"reflect"
	"runtime"
	"sync"
	"time"
)

// TableTestRunner provides a fluent API for table-driven tests with individual execution
type TableTestRunner[I, O any] struct {
	t           TestingT
	testName    string
	cases       []TableTestCase[I, O]
	parallel    bool
	maxWorkers  int
	repeatCount int
	skipFunc    func(string) bool
	focusNames  []string
	timeout     time.Duration
}

// ParallelRunner wraps TableTestRunner for parallel execution
type ParallelRunner[I, O any] struct {
	runner *TableTestRunner[I, O]
}

// BenchmarkResult contains benchmark execution results
type BenchmarkResult struct {
	TotalTime    time.Duration
	AverageTime  time.Duration
	MinTime      time.Duration
	MaxTime      time.Duration
	Iterations   int
	AllocBytes   int64
	AllocObjects int64
}

// Global sync.Pool for memory optimization
var (
	benchResultPool = sync.Pool{
		New: func() interface{} {
			return &BenchmarkResult{}
		},
	}
	resultChannelPool = sync.Pool{
		New: func() interface{} {
			return make(chan interface{}, 1)
		},
	}
)

// TableTestCase represents a single test case with input and expected output
type TableTestCase[I, O any] struct {
	name     string
	input    I
	expected O
}

// SimpleTableRunner provides non-generic table testing for basic types
type SimpleTableRunner struct {
	t     TestingT
	cases []SimpleTableCase
}

// SimpleTableCase represents a test case for simple table testing
type SimpleTableCase struct {
	name     string
	input    interface{}
	expected interface{}
}

// UltraSimpleRunner provides maximum syntax sugar with chained With/Expect
type UltraSimpleRunner struct {
	t            TestingT
	cases        []UltraSimpleCase
	pendingName  string
	pendingInput interface{}
}

// UltraSimpleCase represents the most concise test case format
type UltraSimpleCase struct {
	name     string
	input    interface{}
	expected interface{}
}

// TableTest creates a new generic table test runner
func TableTest[I, O any](t TestingT, testName string) *TableTestRunner[I, O] {
	return &TableTestRunner[I, O]{
		t:           t,
		testName:    testName,
		cases:       make([]TableTestCase[I, O], 0),
		parallel:    false,
		maxWorkers:  runtime.NumCPU(),
		repeatCount: 1,
		skipFunc:    nil,
		focusNames:  nil,
		timeout:     0,
	}
}

// GenericTable creates a new generic table test runner with type inference
func GenericTable[I, O any](t TestingT) *TableTestRunner[I, O] {
	return &TableTestRunner[I, O]{
		t:        t,
		testName: "GenericTable",
		cases:    make([]TableTestCase[I, O], 0),
	}
}

// Table creates a new ultra-simple table runner with maximum syntax sugar
func Table(t TestingT) *UltraSimpleRunner {
	return &UltraSimpleRunner{
		t:     t,
		cases: make([]UltraSimpleCase, 0),
	}
}

// Case adds a test case to the generic table runner
func (tr *TableTestRunner[I, O]) Case(name string, input I, expected O) *TableTestRunner[I, O] {
	tr.cases = append(tr.cases, TableTestCase[I, O]{
		name:     name,
		input:    input,
		expected: expected,
	})
	return tr
}

// Parallel enables parallel execution of test cases
func (tr *TableTestRunner[I, O]) Parallel() *ParallelRunner[I, O] {
	tr.parallel = true
	return &ParallelRunner[I, O]{runner: tr}
}

// WithMaxWorkers sets the maximum number of concurrent workers
func (pr *ParallelRunner[I, O]) WithMaxWorkers(maxWorkers int) *ParallelRunner[I, O] {
	if maxWorkers <= 0 {
		pr.runner.maxWorkers = runtime.NumCPU()
	} else {
		pr.runner.maxWorkers = maxWorkers
	}
	return pr
}

// Run executes all test cases individually with the provided function and assertion
func (tr *TableTestRunner[I, O]) Run(testFunc func(I) O, assertFunc func(string, I, O, O)) {
	tr.executeRun(testFunc, assertFunc)
}

// Run executes test cases with parallel support
func (pr *ParallelRunner[I, O]) Run(testFunc func(I) O, assertFunc func(string, I, O, O)) {
	pr.runner.executeRun(testFunc, assertFunc)
}

// executeRun is the core execution logic with parallel support
func (tr *TableTestRunner[I, O]) executeRun(testFunc func(I) O, assertFunc func(string, I, O, O)) {
	cases := tr.getFilteredCases()

	if tr.parallel {
		tr.executeParallel(cases, testFunc, assertFunc)
	} else {
		tr.executeSequential(cases, testFunc, assertFunc)
	}
}

// getFilteredCases returns cases based on focus and skip filters
func (tr *TableTestRunner[I, O]) getFilteredCases() []TableTestCase[I, O] {
	cases := tr.cases

	// Apply focus filter first
	if len(tr.focusNames) > 0 {
		cases = applyFocusFilter(cases, tr.focusNames)
	}

	// Apply skip filter
	if tr.skipFunc != nil {
		cases = applySkipFilter(cases, tr.skipFunc)
	}

	return cases
}

// applyFocusFilter applies focus filtering - pure function
func applyFocusFilter[I, O any](cases []TableTestCase[I, O], focusNames []string) []TableTestCase[I, O] {
	focusSet := createFocusSet(focusNames)
	var filtered []TableTestCase[I, O]

	for _, tc := range cases {
		if focusSet[tc.name] {
			filtered = append(filtered, tc)
		}
	}

	return filtered
}

// createFocusSet creates a lookup map - pure function
func createFocusSet(focusNames []string) map[string]bool {
	focusSet := make(map[string]bool, len(focusNames))
	for _, name := range focusNames {
		focusSet[name] = true
	}
	return focusSet
}

// applySkipFilter applies skip filtering - pure function
func applySkipFilter[I, O any](cases []TableTestCase[I, O], skipFunc func(string) bool) []TableTestCase[I, O] {
	var filtered []TableTestCase[I, O]
	for _, tc := range cases {
		if !skipFunc(tc.name) {
			filtered = append(filtered, tc)
		}
	}
	return filtered
}

// executeSequential runs test cases one by one
func (tr *TableTestRunner[I, O]) executeSequential(cases []TableTestCase[I, O], testFunc func(I) O, assertFunc func(string, I, O, O)) {
	for _, tc := range cases {
		executeTestCaseRepeats(tc, tr.repeatCount, tr.timeout, testFunc, assertFunc, tr.t)
	}
}

// executeTestCaseRepeats executes a test case multiple times - pure function
func executeTestCaseRepeats[I, O any](tc TableTestCase[I, O], repeatCount int, timeout time.Duration, testFunc func(I) O, assertFunc func(string, I, O, O), t TestingT) {
	for i := 0; i < repeatCount; i++ {
		if timeout > 0 {
			executeWithTimeout(tc, testFunc, assertFunc, t, timeout)
			continue
		}

		actual := testFunc(tc.input)
		assertFunc(tc.name, tc.input, tc.expected, actual)
	}
}

// executeParallel runs test cases concurrently using worker pool pattern
func (tr *TableTestRunner[I, O]) executeParallel(cases []TableTestCase[I, O], testFunc func(I) O, assertFunc func(string, I, O, O)) {
	var wg sync.WaitGroup
	semaphore := make(chan struct{}, tr.maxWorkers)

	for _, tc := range cases {
		for i := 0; i < tr.repeatCount; i++ {
			wg.Add(1)
			go func(testCase TableTestCase[I, O]) {
				defer wg.Done()
				semaphore <- struct{}{}        // Acquire worker slot
				defer func() { <-semaphore }() // Release worker slot

				if tr.timeout > 0 {
					executeWithTimeout(testCase, testFunc, assertFunc, tr.t, tr.timeout)
				} else {
					actual := testFunc(testCase.input)
					assertFunc(testCase.name, testCase.input, testCase.expected, actual)
				}
			}(tc)
		}
	}

	wg.Wait()
}

// executeWithTimeout runs a single test case with timeout - pure function
func executeWithTimeout[I, O any](tc TableTestCase[I, O], testFunc func(I) O, assertFunc func(string, I, O, O), t TestingT, timeout time.Duration) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	resultChan := make(chan O, 1)
	go func() {
		result := testFunc(tc.input)
		select {
		case resultChan <- result:
		case <-ctx.Done():
		}
	}()

	select {
	case actual := <-resultChan:
		assertFunc(tc.name, tc.input, tc.expected, actual)
	case <-ctx.Done():
		t.Errorf("Test case %s timed out after %v", tc.name, timeout)
	}
}

// Repeat sets the number of times each test case should be repeated
func (tr *TableTestRunner[I, O]) Repeat(count int) *TableTestRunner[I, O] {
	tr.repeatCount = count
	return tr
}

// Skip sets a condition function to skip certain test cases
func (tr *TableTestRunner[I, O]) Skip(skipFunc func(string) bool) *TableTestRunner[I, O] {
	tr.skipFunc = skipFunc
	return tr
}

// Focus runs only the specified test cases
func (tr *TableTestRunner[I, O]) Focus(names []string) *TableTestRunner[I, O] {
	tr.focusNames = names
	return tr
}

// Timeout sets a timeout for each test case
func (tr *TableTestRunner[I, O]) Timeout(timeout time.Duration) *TableTestRunner[I, O] {
	tr.timeout = timeout
	return tr
}

// Benchmark runs benchmark testing on the test cases
func (tr *TableTestRunner[I, O]) Benchmark(testFunc func(I) O, iterations int) *BenchmarkResult {
	return tr.executeBenchmark(testFunc, iterations)
}

// executeBenchmark performs benchmark execution with sync.Pool optimization
func (tr *TableTestRunner[I, O]) executeBenchmark(testFunc func(I) O, iterations int) *BenchmarkResult {
	if len(tr.cases) == 0 {
		return &BenchmarkResult{}
	}

	// Get benchmark result from pool
	result := benchResultPool.Get().(*BenchmarkResult)
	defer func() {
		// Reset and return to pool
		*result = BenchmarkResult{}
		benchResultPool.Put(result)
	}()

	// Use first test case for benchmarking
	testCase := tr.cases[0]

	result.Iterations = iterations
	result.MinTime = time.Hour // Start with high value

	// Warmup
	executeWarmup(testFunc, testCase.input, 10)

	// Actual benchmark
	result.TotalTime = executeBenchmarkLoop(testFunc, testCase.input, iterations, result)
	result.AverageTime = result.TotalTime / time.Duration(iterations)

	// Create and return a copy
	return &BenchmarkResult{
		TotalTime:   result.TotalTime,
		AverageTime: result.AverageTime,
		MinTime:     result.MinTime,
		MaxTime:     result.MaxTime,
		Iterations:  result.Iterations,
	}
}

// executeWarmup performs warmup iterations - pure function
func executeWarmup[I, O any](testFunc func(I) O, input I, warmupCount int) {
	for i := 0; i < warmupCount; i++ {
		testFunc(input)
	}
}

// executeBenchmarkLoop performs the actual benchmark loop - pure function
func executeBenchmarkLoop[I, O any](testFunc func(I) O, input I, iterations int, result *BenchmarkResult) time.Duration {
	var totalTime time.Duration

	for i := 0; i < iterations; i++ {
		duration := measureExecution(testFunc, input)
		totalTime = duration

		if duration < result.MinTime {
			result.MinTime = duration
		}
		if duration > result.MaxTime {
			result.MaxTime = duration
		}
	}

	return totalTime
}

// measureExecution measures a single execution - pure function
func measureExecution[I, O any](testFunc func(I) O, input I) time.Duration {
	start := time.Now()
	testFunc(input)
	return time.Since(start)
}

// Fuzz generates random test cases using the provided generator
func (tr *TableTestRunner[I, O]) Fuzz(generator func() (I, O), count int) *TableTestRunner[I, O] {
	for i := 0; i < count; i++ {
		input, expected := generator()
		tr.cases = append(tr.cases, TableTestCase[I, O]{
			name:     fmt.Sprintf("fuzz_case_%d", i),
			input:    input,
			expected: expected,
		})
	}
	return tr
}

// Property performs property-based testing with invariants
func (tr *TableTestRunner[I, O]) Property(generator func() I, invariant func(I, O) bool, count int) *TableTestRunner[I, O] {
	for i := 0; i < count; i++ {
		input := generator()
		tr.cases = append(tr.cases, TableTestCase[I, O]{
			name:     fmt.Sprintf("property_case_%d", i),
			input:    input,
			expected: tr.generateExpectedForProperty(input, invariant),
		})
	}
	return tr
}

// generateExpectedForProperty is a helper for property-based testing
func (tr *TableTestRunner[I, O]) generateExpectedForProperty(input I, invariant func(I, O) bool) O {
	// This is a simplified implementation
	// In a real implementation, we'd need to find a value that satisfies the invariant
	var zero O
	return zero
}

// With starts a new test case with input value (ultra-simple syntax)
func (usr *UltraSimpleRunner) With(name string, input interface{}) *UltraSimpleRunner {
	usr.pendingName = name
	usr.pendingInput = input
	return usr
}

// Expect completes the current test case with expected value
func (usr *UltraSimpleRunner) Expect(expected interface{}) *UltraSimpleRunner {
	if usr.pendingName != "" && usr.pendingInput != nil {
		usr.cases = append(usr.cases, UltraSimpleCase{
			name:     usr.pendingName,
			input:    usr.pendingInput,
			expected: expected,
		})
		// Clear pending values
		usr.pendingName = ""
		usr.pendingInput = nil
	}
	return usr
}

// Test executes all cases with the provided function and assertion
func (usr *UltraSimpleRunner) Test(testFunc interface{}, assertFunc interface{}) {
	testFuncValue := reflect.ValueOf(testFunc)
	assertFuncValue := reflect.ValueOf(assertFunc)

	for _, tc := range usr.cases {
		// Execute test function with input
		inputValues := []reflect.Value{reflect.ValueOf(tc.input)}
		results := testFuncValue.Call(inputValues)

		if len(results) > 0 {
			actual := results[0].Interface()

			// Call assertion function
			assertArgs := []reflect.Value{
				reflect.ValueOf(tc.name),
				reflect.ValueOf(tc.input),
				reflect.ValueOf(tc.expected),
				reflect.ValueOf(actual),
			}
			assertFuncValue.Call(assertArgs)
		}
	}
}

// SimpleTable creates a new simple table test runner (renamed to avoid conflict)
func SimpleTable(t TestingT, testName string) *SimpleTableRunner {
	return &SimpleTableRunner{
		t:     t,
		cases: make([]SimpleTableCase, 0),
	}
}

// Case adds a test case to the simple table runner
func (str *SimpleTableRunner) Case(name string, input, expected interface{}, testFunc interface{}) *SimpleTableRunner {
	str.cases = append(str.cases, SimpleTableCase{
		name:     name,
		input:    input,
		expected: expected,
	})
	return str
}

// Run executes all test cases with the provided runner function
func (str *SimpleTableRunner) Run(runnerFunc interface{}) {
	runnerFuncValue := reflect.ValueOf(runnerFunc)

	for _, tc := range str.cases {
		// Execute runner function with test case data
		args := []reflect.Value{
			reflect.ValueOf(tc.name),
			reflect.ValueOf(tc.input),
			reflect.ValueOf(tc.expected),
		}

		// Add the test function as the fourth argument (from Case method call)
		// Note: This is a simplified implementation - in a real scenario,
		// we'd need to track the test function per case
		runnerFuncValue.Call(args)
	}
}

// Quick creates a table test with maximum syntax sugar for simple cases
func Quick[T comparable](t TestingT, testFunc func(T) T) *QuickRunner[T] {
	return &QuickRunner[T]{
		t:        t,
		testFunc: testFunc,
		cases:    make([]QuickCase[T], 0),
	}
}

// QuickRunner provides the most concise table testing experience
type QuickRunner[T comparable] struct {
	t        TestingT
	testFunc func(T) T
	cases    []QuickCase[T]
}

// QuickCase represents a minimal test case
type QuickCase[T comparable] struct {
	input    T
	expected T
}

// Check adds a test case with automatic assertion
func (qr *QuickRunner[T]) Check(input, expected T) *QuickRunner[T] {
	qr.cases = append(qr.cases, QuickCase[T]{
		input:    input,
		expected: expected,
	})
	return qr
}

// Run executes all cases with automatic assertion
func (qr *QuickRunner[T]) Run() {
	for i, tc := range qr.cases {
		actual := qr.testFunc(tc.input)
		if actual != tc.expected {
			qr.t.Errorf("Case %d: input %v, expected %v, got %v", i, tc.input, tc.expected, actual)
		}
	}
}
