// Package vasdeference - Advanced Table Testing Framework Example
// This file demonstrates all the advanced features implemented following TDD methodology
package vasdeference

import (
	"fmt"
	"math/rand"
	"time"
)

// ExampleAdvancedTableTesting demonstrates all advanced features
func ExampleAdvancedTableTesting() {
	// Mock testing interface for demonstration
	mockT := &mockTestingT{}
	
	// =================================================================
	// 1. PARALLEL EXECUTION WITH WORKER POOL
	// =================================================================
	fmt.Println("=== Parallel Execution Example ===")
	
	parallelRunner := TableTest[int, int](mockT, "parallel_math_operations").
		Case("square_1", 1, 1).
		Case("square_2", 2, 4).
		Case("square_3", 3, 9).
		Case("square_4", 4, 16).
		Case("square_5", 5, 25)
	
	mathFunc := func(x int) int {
		// Simulate some computation time
		time.Sleep(10 * time.Millisecond)
		return x * x
	}
	
	// Execute with controlled concurrency
	start := time.Now()
	parallelRunner.Parallel().WithMaxWorkers(3).Run(mathFunc, func(name string, input, expected, actual int) {
		fmt.Printf("✓ %s: %d² = %d\n", name, input, actual)
	})
	parallelTime := time.Since(start)
	fmt.Printf("Parallel execution completed in: %v\n\n", parallelTime)
	
	// =================================================================
	// 2. BENCHMARKING INTEGRATION
	// =================================================================
	fmt.Println("=== Benchmarking Example ===")
	
	benchRunner := TableTest[int, int](mockT, "fibonacci_benchmark").
		Case("fib_10", 10, fibonacci(10)).
		Case("fib_20", 20, fibonacci(20)).
		Case("fib_25", 25, fibonacci(25))
	
	// Run benchmark
	benchResult := benchRunner.Benchmark(fibonacci, 1000)
	fmt.Printf("Benchmark Results:\n")
	fmt.Printf("  Total Time: %v\n", benchResult.TotalTime)
	fmt.Printf("  Average: %v\n", benchResult.AverageTime)
	fmt.Printf("  Min: %v, Max: %v\n", benchResult.MinTime, benchResult.MaxTime)
	fmt.Printf("  Iterations: %d\n\n", benchResult.Iterations)
	
	// =================================================================
	// 3. ADVANCED SYNTAX SUGAR - CHAINING
	// =================================================================
	fmt.Println("=== Advanced Method Chaining Example ===")
	
	chainRunner := TableTest[string, int](mockT, "string_processing").
		Case("hello", "hello", 5).
		Case("world", "world", 5).
		Case("go", "go", 2).
		Case("programming", "programming", 11).
		Case("skip_me", "skip_me", 7). // This will be skipped
		Skip(func(name string) bool { 
			return name == "skip_me" 
		}).
		Focus([]string{"hello", "world", "programming"}). // Focus on specific cases
		Repeat(2). // Execute each case twice
		Timeout(100 * time.Millisecond) // Set timeout
	
	stringLengthFunc := func(s string) int {
		time.Sleep(5 * time.Millisecond) // Simulate processing
		return len(s)
	}
	
	chainRunner.Parallel().WithMaxWorkers(2).Run(stringLengthFunc, func(name string, input string, expected, actual int) {
		fmt.Printf("✓ %s: len('%s') = %d\n", name, input, actual)
	})
	fmt.Println()
	
	// =================================================================
	// 4. FUZZING SUPPORT
	// =================================================================
	fmt.Println("=== Fuzzing Example ===")
	
	fuzzRunner := TableTest[int, bool](mockT, "even_number_fuzz")
	
	// Generate random test cases
	fuzzRunner.Fuzz(func() (int, bool) {
		input := rand.Intn(100)
		expected := input%2 == 0
		return input, expected
	}, 10) // Generate 10 random test cases
	
	isEvenFunc := func(x int) bool {
		return x%2 == 0
	}
	
	fuzzRunner.Run(isEvenFunc, func(name string, input int, expected, actual bool) {
		status := "✓"
		if expected != actual {
			status = "✗"
		}
		fmt.Printf("%s %s: isEven(%d) = %t\n", status, name, input, actual)
	})
	fmt.Println()
	
	// =================================================================
	// 5. PROPERTY-BASED TESTING
	// =================================================================
	fmt.Println("=== Property-Based Testing Example ===")
	
	propertyRunner := TableTest[int, int](mockT, "addition_properties")
	
	// Test commutative property: a + b = b + a
	propertyRunner.Property(
		func() int { return rand.Intn(50) }, // Generate random integers
		func(input, output int) bool {
			// Property: adding 0 should return the same number
			return input+0 == output
		},
		5, // Generate 5 test cases
	)
	
	addZeroFunc := func(x int) int {
		return x + 0
	}
	
	propertyRunner.Run(addZeroFunc, func(name string, input, expected, actual int) {
		fmt.Printf("✓ %s: %d + 0 = %d\n", name, input, actual)
	})
	fmt.Println()
	
	// =================================================================
	// 6. MEMORY OPTIMIZATION WITH SYNC.POOL
	// =================================================================
	fmt.Println("=== Memory Pool Optimization Example ===")
	
	// Create multiple runners to demonstrate pool reuse
	for i := 0; i < 5; i++ {
		poolRunner := TableTest[int, int](mockT, fmt.Sprintf("pool_test_%d", i)).
			Case("computation", i*10, i*10*2)
		
		// Each benchmark will reuse pooled objects
		result := poolRunner.Benchmark(func(x int) int {
			sum := 0
			for j := 0; j < x; j++ {
				sum += j
			}
			return sum
		}, 100)
		
		fmt.Printf("Pool test %d: avg %v\n", i, result.AverageTime)
	}
	fmt.Println()
	
	fmt.Println("=== All Advanced Features Demonstrated ===")
}

// fibonacci calculates fibonacci number (for benchmarking example)
func fibonacci(n int) int {
	if n <= 1 {
		return n
	}
	return fibonacci(n-1) + fibonacci(n-2)
}

// ExampleQuickSyntax demonstrates the ultra-simple syntax
func ExampleQuickSyntax() {
	mockT := &mockTestingT{}
	
	// Ultra-simple syntax for quick testing
	Quick(mockT, func(x int) int { return x * 2 }).
		Check(1, 2).
		Check(3, 6).
		Check(5, 10).
		Run()
	
	fmt.Println("✓ Quick syntax example completed")
}

// ExampleUltraSimpleSyntax demonstrates the most concise syntax
func ExampleUltraSimpleSyntax() {
	mockT := &mockTestingT{}
	
	// Ultra-simple chained syntax
	Table(mockT).
		With("double 2", 2).Expect(4).
		With("double 4", 4).Expect(8).
		With("double 6", 6).Expect(12).
		Test(func(x int) int { return x * 2 }, func(name string, input, expected, actual interface{}) {
			fmt.Printf("✓ %s: %v * 2 = %v\n", name, input, actual)
		})
	
	fmt.Println("✓ Ultra-simple syntax example completed")
}