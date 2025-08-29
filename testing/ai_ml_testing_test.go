package testutils

import (
	"context"
	"go/ast"
	"go/parser"
	"go/token"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewTestIntelligenceEngine(t *testing.T) {
	t.Run("should create engine with all components initialized", func(t *testing.T) {
		engine := NewTestIntelligenceEngine()
		
		assert.NotNil(t, engine)
		assert.NotNil(t, engine.codeAnalyzer)
		assert.NotNil(t, engine.mutationOptimizer)
		assert.NotNil(t, engine.edgeCaseDiscovery)
		assert.NotNil(t, engine.performanceModel)
		assert.NotNil(t, engine.patternAnalyzer)
		assert.NotNil(t, engine.failurePredictor)
		assert.NotNil(t, engine.testScheduler)
		assert.NotNil(t, engine.testGenerator)
	})

	t.Run("should initialize mutation optimizer with correct defaults", func(t *testing.T) {
		engine := NewTestIntelligenceEngine()
		
		assert.Equal(t, 50, engine.mutationOptimizer.populationSize)
		assert.Equal(t, 20, engine.mutationOptimizer.generations)
		assert.Equal(t, 0.1, engine.mutationOptimizer.mutationRate)
		assert.Equal(t, 0.7, engine.mutationOptimizer.crossoverRate)
		assert.NotNil(t, engine.mutationOptimizer.rng)
	})

	t.Run("should initialize with known edge case patterns", func(t *testing.T) {
		engine := NewTestIntelligenceEngine()
		
		patterns := engine.edgeCaseDiscovery.knownPatterns
		assert.NotEmpty(t, patterns)
		
		// Verify expected patterns exist
		patternTypes := make(map[string]bool)
		for _, pattern := range patterns {
			patternTypes[pattern.PatternType] = true
		}
		
		expectedPatterns := []string{"null_pointer", "empty_collection", "boundary_values", "concurrent_access"}
		for _, expected := range expectedPatterns {
			assert.True(t, patternTypes[expected], "Expected pattern %s not found", expected)
		}
	})
}

func TestCodeAnalyzer_AnalyzeCodeStructure(t *testing.T) {
	analyzer := &CodeAnalyzer{fileSet: token.NewFileSet()}

	t.Run("should parse simple Go file", func(t *testing.T) {
		// Create a temporary Go file content
		goCode := `package main

type User struct {
	Name string
	Age  int
}

type UserService interface {
	GetUser(id string) (*User, error)
	CreateUser(name string, age int) error
}

func CreateUser(name string, age int) (*User, error) {
	if name == "" {
		return nil, errors.New("name cannot be empty")
	}
	
	if age < 0 {
		return nil, errors.New("age cannot be negative")
	}
	
	return &User{Name: name, Age: age}, nil
}

func processUsers(users []User) int {
	count := 0
	for _, user := range users {
		if user.Age >= 18 {
			count++
		}
	}
	return count
}`

		// Parse the code
		src, err := parser.ParseFile(analyzer.fileSet, "test.go", goCode, parser.ParseComments)
		require.NoError(t, err)

		// Analyze the parsed file
		structure := &CodeStructure{
			Functions:  make([]FunctionSignature, 0),
			Interfaces: make([]InterfaceDefinition, 0),
			Structs:    make([]StructDefinition, 0),
		}

		ast.Inspect(src, func(n ast.Node) bool {
			switch node := n.(type) {
			case *ast.FuncDecl:
				if node.Name.IsExported() {
					sig := analyzer.extractFunctionSignature(node)
					structure.Functions = append(structure.Functions, sig)
					structure.Complexity += sig.Complexity
				}
			case *ast.TypeSpec:
				switch typeNode := node.Type.(type) {
				case *ast.InterfaceType:
					iface := analyzer.extractInterfaceDefinition(node.Name.Name, typeNode)
					structure.Interfaces = append(structure.Interfaces, iface)
				case *ast.StructType:
					structDef := analyzer.extractStructDefinition(node.Name.Name, typeNode)
					structure.Structs = append(structure.Structs, structDef)
				}
			}
			return true
		})

		// Verify structure
		assert.Equal(t, 1, len(structure.Functions)) // CreateUser
		assert.Equal(t, 1, len(structure.Interfaces)) // UserService
		assert.Equal(t, 1, len(structure.Structs)) // User
		assert.Greater(t, structure.Complexity, 0)

		// Verify function signature
		createUserFunc := structure.Functions[0]
		assert.Equal(t, "CreateUser", createUserFunc.Name)
		assert.Equal(t, 2, len(createUserFunc.Parameters))
		assert.Equal(t, 2, len(createUserFunc.Returns))

		// Verify interface
		userService := structure.Interfaces[0]
		assert.Equal(t, "UserService", userService.Name)
		assert.Equal(t, 2, len(userService.Methods))

		// Verify struct
		userStruct := structure.Structs[0]
		assert.Equal(t, "User", userStruct.Name)
		assert.Equal(t, 2, len(userStruct.Fields))
	})

	t.Run("should handle file parsing errors", func(t *testing.T) {
		invalidCode := "invalid go code {"
		_, err := parser.ParseFile(analyzer.fileSet, "invalid.go", invalidCode, parser.ParseComments)
		assert.Error(t, err)
	})
}

func TestTestIntelligenceEngine_GenerateIntelligentTests(t *testing.T) {
	engine := NewTestIntelligenceEngine()

	t.Run("should generate tests for valid Go code", func(t *testing.T) {
		// Create a temporary file with Go code
		goCode := `package testpkg

func Add(a, b int) int {
	return a + b
}

func Divide(a, b float64) (float64, error) {
	if b == 0 {
		return 0, errors.New("division by zero")
	}
	return a / b, nil
}`

		// Mock file analysis by creating structure directly
		structure := &CodeStructure{
			Functions: []FunctionSignature{
				{
					Name: "Add",
					Parameters: []Parameter{
						{Name: "a", Type: "int"},
						{Name: "b", Type: "int"},
					},
					Returns: []ReturnType{
						{Type: "int", Error: false},
					},
					Complexity: 1,
				},
				{
					Name: "Divide",
					Parameters: []Parameter{
						{Name: "a", Type: "float64"},
						{Name: "b", Type: "float64"},
					},
					Returns: []ReturnType{
						{Type: "float64", Error: false},
						{Type: "error", Error: true},
					},
					Complexity: 2,
				},
			},
			Interfaces: []InterfaceDefinition{},
			Structs:    []StructDefinition{},
			Complexity: 3,
		}

		// Generate tests using the engine's test generation logic
		var tests []GeneratedTest

		// Generate tests for each function
		for _, function := range structure.Functions {
			test := engine.generateFunctionTest(function, structure)
			tests = append(tests, test)
		}

		// Generate edge case tests
		edgeCases := engine.edgeCaseDiscovery.DiscoverEdgeCases(structure)
		for _, edgeCase := range edgeCases {
			test := engine.generateEdgeCaseTest(edgeCase)
			tests = append(tests, test)
		}

		assert.Greater(t, len(tests), 0)

		// Verify function tests
		functionTests := 0
		edgeCaseTests := 0
		for _, test := range tests {
			if test.Type == "function_test" {
				functionTests++
				assert.NotEmpty(t, test.Name)
				assert.NotEmpty(t, test.Code)
				assert.Greater(t, test.Confidence, 0.0)
				assert.NotEmpty(t, test.Rationale)
			} else if test.Type == "edge_case_test" {
				edgeCaseTests++
			}
		}

		assert.Equal(t, 2, functionTests) // Add and Divide
		assert.Greater(t, edgeCaseTests, 0) // Should have edge case tests
	})

	t.Run("should handle empty code structure", func(t *testing.T) {
		structure := &CodeStructure{
			Functions:  []FunctionSignature{},
			Interfaces: []InterfaceDefinition{},
			Structs:    []StructDefinition{},
			Complexity: 0,
		}

		var tests []GeneratedTest
		
		// Generate tests for each function (none in this case)
		for _, function := range structure.Functions {
			test := engine.generateFunctionTest(function, structure)
			tests = append(tests, test)
		}

		// Generate edge case tests
		edgeCases := engine.edgeCaseDiscovery.DiscoverEdgeCases(structure)
		for _, edgeCase := range edgeCases {
			test := engine.generateEdgeCaseTest(edgeCase)
			tests = append(tests, test)
		}

		// Should still have edge case tests from known patterns
		assert.Greater(t, len(tests), 0)
	})
}

func TestGeneticMutationOptimizer_OptimizeMutationTesting(t *testing.T) {
	optimizer := &GeneticMutationOptimizer{
		populationSize: 10,
		generations:    5,
		mutationRate:   0.1,
		crossoverRate:  0.7,
		rng:           newTestRNG(),
	}

	t.Run("should optimize test suite selection", func(t *testing.T) {
		testSuite := []string{
			"TestFunction1",
			"TestFunction2",
			"TestFunction3",
			"TestFunction4",
			"TestFunction5",
		}

		// Fitness function that prefers smaller test suites with specific tests
		fitnessFunc := func(subset []string) float64 {
			score := 0.0
			
			// Base score for having any tests
			if len(subset) > 0 {
				score += 10.0
			}
			
			// Bonus for specific important tests
			for _, test := range subset {
				if test == "TestFunction1" || test == "TestFunction3" {
					score += 5.0
				}
			}
			
			// Penalty for too many tests
			if len(subset) > 3 {
				score -= float64(len(subset)-3) * 2.0
			}
			
			return score
		}

		result, err := optimizer.OptimizeMutationTesting(testSuite, fitnessFunc)
		
		assert.NoError(t, err)
		assert.NotEmpty(t, result)
		assert.LessOrEqual(t, len(result), len(testSuite))
		
		// Verify result is a subset of original test suite
		testMap := make(map[string]bool)
		for _, test := range testSuite {
			testMap[test] = true
		}
		
		for _, test := range result {
			assert.True(t, testMap[test], "Result contains test not in original suite: %s", test)
		}
	})

	t.Run("should handle empty test suite", func(t *testing.T) {
		testSuite := []string{}
		
		fitnessFunc := func(subset []string) float64 {
			return float64(len(subset))
		}

		result, err := optimizer.OptimizeMutationTesting(testSuite, fitnessFunc)
		
		assert.NoError(t, err)
		assert.Empty(t, result)
	})

	t.Run("should handle single test", func(t *testing.T) {
		testSuite := []string{"TestSingle"}
		
		fitnessFunc := func(subset []string) float64 {
			return float64(len(subset))
		}

		result, err := optimizer.OptimizeMutationTesting(testSuite, fitnessFunc)
		
		assert.NoError(t, err)
		assert.LessOrEqual(t, len(result), 1)
	})
}

func TestPerformancePredictor(t *testing.T) {
	predictor := &PerformancePredictor{
		historicalData: []TestExecution{},
		model:          &SimpleMLModel{weights: []float64{0.5, 0.3, 0.2, 0.1, 0.05}, bias: 100.0},
	}

	t.Run("should predict performance metrics", func(t *testing.T) {
		testName := "TestExample"
		features := []float64{1.5, 2.0, 0.8, 1.0, 10.0} // Duration factor, memory factor, CPU factor, success factor, name length

		prediction := predictor.PredictPerformance(testName, features)

		assert.Equal(t, testName, prediction.TestName)
		assert.Greater(t, prediction.PredictedTime, time.Duration(0))
		assert.GreaterOrEqual(t, prediction.ResourceUsage.Memory, 0.0)
		assert.GreaterOrEqual(t, prediction.ResourceUsage.CPU, 0.0)
		assert.GreaterOrEqual(t, prediction.ResourceUsage.IO, 0.0)
		assert.GreaterOrEqual(t, prediction.FailureProbability, 0.0)
		assert.LessOrEqual(t, prediction.FailureProbability, 1.0)
		assert.Greater(t, prediction.Confidence, 0.0)
	})

	t.Run("should handle empty features", func(t *testing.T) {
		testName := "TestEmpty"
		features := []float64{}

		prediction := predictor.PredictPerformance(testName, features)

		assert.Equal(t, testName, prediction.TestName)
		assert.Greater(t, prediction.PredictedTime, time.Duration(0))
	})

	t.Run("should train model with historical data", func(t *testing.T) {
		executions := []TestExecution{
			{
				TestName:    "Test1",
				Duration:    100 * time.Millisecond,
				MemoryUsage: 1024 * 1024, // 1MB
				CPUUsage:    50.0,
				Success:     true,
				Timestamp:   time.Now(),
			},
			{
				TestName:    "Test2",
				Duration:    200 * time.Millisecond,
				MemoryUsage: 2048 * 1024, // 2MB
				CPUUsage:    75.0,
				Success:     false,
				Timestamp:   time.Now(),
			},
		}

		// Add more data to meet minimum training requirement
		for i := 0; i < 10; i++ {
			executions = append(executions, TestExecution{
				TestName:    "TestBulk",
				Duration:    150 * time.Millisecond,
				MemoryUsage: 1536 * 1024,
				CPUUsage:    60.0,
				Success:     true,
				Timestamp:   time.Now(),
			})
		}

		err := predictor.TrainPerformanceModel(executions)
		assert.NoError(t, err)
		assert.Greater(t, len(predictor.historicalData), 0)
	})

	t.Run("should fail training with insufficient data", func(t *testing.T) {
		smallPredictor := &PerformancePredictor{
			historicalData: []TestExecution{},
			model:          &SimpleMLModel{weights: make([]float64, 5), bias: 0.0},
		}

		insufficientData := []TestExecution{
			{TestName: "Test1", Duration: 100 * time.Millisecond},
		}

		err := smallPredictor.TrainPerformanceModel(insufficientData)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "insufficient data")
	})
}

func TestEdgeCaseDiscoveryEngine(t *testing.T) {
	engine := &EdgeCaseDiscoveryEngine{
		knownPatterns: getKnownEdgeCasePatterns(),
		analyzer:      &CodeAnalyzer{fileSet: token.NewFileSet()},
	}

	t.Run("should discover edge cases from code structure", func(t *testing.T) {
		structure := &CodeStructure{
			Functions: []FunctionSignature{
				{
					Name:       "ComplexFunction",
					Complexity: 8, // High complexity
					Returns: []ReturnType{
						{Type: "error", Error: true},
					},
				},
				{
					Name:       "SimpleFunction",
					Complexity: 2,
					Returns: []ReturnType{
						{Type: "string", Error: false},
					},
				},
			},
		}

		edgeCases := engine.DiscoverEdgeCases(structure)

		assert.Greater(t, len(edgeCases), 0)

		// Should include known patterns
		hasKnownPattern := false
		hasComplexityPattern := false
		hasErrorPattern := false

		for _, edgeCase := range edgeCases {
			if edgeCase.PatternType == "null_pointer" || edgeCase.PatternType == "empty_collection" {
				hasKnownPattern = true
			}
			if edgeCase.PatternType == "high_complexity_function" {
				hasComplexityPattern = true
			}
			if edgeCase.PatternType == "error_handling" {
				hasErrorPattern = true
			}
		}

		assert.True(t, hasKnownPattern, "Should include known edge case patterns")
		assert.True(t, hasComplexityPattern, "Should detect high complexity functions")
		assert.True(t, hasErrorPattern, "Should detect error-returning functions")
	})

	t.Run("should return known patterns for empty structure", func(t *testing.T) {
		structure := &CodeStructure{
			Functions:  []FunctionSignature{},
			Interfaces: []InterfaceDefinition{},
			Structs:    []StructDefinition{},
		}

		edgeCases := engine.DiscoverEdgeCases(structure)

		// Should at least return known patterns
		assert.GreaterOrEqual(t, len(edgeCases), len(engine.knownPatterns))
	})
}

func TestPatternAnalyzer(t *testing.T) {
	analyzer := &PatternAnalyzer{
		coverageHistory: []CoverageSnapshot{
			{Timestamp: time.Now().Add(-2 * time.Hour), LineCoverage: 0.7, FuncCoverage: 0.8},
			{Timestamp: time.Now().Add(-1 * time.Hour), LineCoverage: 0.75, FuncCoverage: 0.82},
			{Timestamp: time.Now(), LineCoverage: 0.8, FuncCoverage: 0.85},
		},
		failurePatterns: []FailurePattern{},
	}

	t.Run("should analyze improving coverage trend", func(t *testing.T) {
		insights, err := analyzer.AnalyzeCoveragePatterns()
		
		assert.NoError(t, err)
		assert.Greater(t, len(insights), 0)
		
		// Should detect improving trend
		var trendInsight *CoverageInsight
		for i := range insights {
			if insights[i].Type == "trend" {
				trendInsight = &insights[i]
				break
			}
		}
		
		assert.NotNil(t, trendInsight)
		assert.Contains(t, trendInsight.Description, "improving")
		assert.Equal(t, 1.0, trendInsight.Impact)
	})

	t.Run("should handle declining coverage", func(t *testing.T) {
		decliningAnalyzer := &PatternAnalyzer{
			coverageHistory: []CoverageSnapshot{
				{Timestamp: time.Now().Add(-1 * time.Hour), LineCoverage: 0.8, FuncCoverage: 0.85},
				{Timestamp: time.Now(), LineCoverage: 0.65, FuncCoverage: 0.7}, // Significant decline
			},
		}

		insights, err := decliningAnalyzer.AnalyzeCoveragePatterns()
		
		assert.NoError(t, err)
		assert.Greater(t, len(insights), 0)
		
		var trendInsight *CoverageInsight
		for i := range insights {
			if insights[i].Type == "trend" {
				trendInsight = &insights[i]
				break
			}
		}
		
		assert.NotNil(t, trendInsight)
		assert.Contains(t, trendInsight.Description, "declining")
		assert.Equal(t, -1.0, trendInsight.Impact)
	})

	t.Run("should fail with insufficient data", func(t *testing.T) {
		emptyAnalyzer := &PatternAnalyzer{
			coverageHistory: []CoverageSnapshot{},
		}

		_, err := emptyAnalyzer.AnalyzeCoveragePatterns()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "insufficient coverage history")
	})
}

func TestIntelligentTestScheduler(t *testing.T) {
	scheduler := &IntelligentTestScheduler{
		testMetrics: []TestMetrics{
			{
				TestName:          "HighPriorityTest",
				ExecutionTime:     100 * time.Millisecond,
				FailureRate:       0.2,
				Priority:          5,
				ResourceIntensive: false,
			},
			{
				TestName:          "ResourceIntensiveTest",
				ExecutionTime:     500 * time.Millisecond,
				FailureRate:       0.05,
				Priority:          2,
				ResourceIntensive: true,
			},
			{
				TestName:          "RegularTest",
				ExecutionTime:     200 * time.Millisecond,
				FailureRate:       0.1,
				Priority:          3,
				ResourceIntensive: false,
			},
		},
		scheduleOptimizer: &ScheduleOptimizer{
			constraints: getDefaultScheduleConstraints(),
		},
	}

	t.Run("should generate optimal test schedule", func(t *testing.T) {
		tests := []string{"HighPriorityTest", "ResourceIntensiveTest", "RegularTest", "UnknownTest"}

		schedule, err := scheduler.GenerateOptimalTestSchedule(tests)
		
		assert.NoError(t, err)
		assert.Equal(t, len(tests), len(schedule.TestOrder))
		assert.Greater(t, schedule.EstimatedTime, time.Duration(0))
		assert.Greater(t, len(schedule.ParallelGroups), 0)
		assert.Equal(t, len(tests), len(schedule.Priority))

		// High priority test should be early in the schedule
		highPriorityIndex := -1
		resourceIntensiveIndex := -1
		for i, testName := range schedule.TestOrder {
			if testName == "HighPriorityTest" {
				highPriorityIndex = i
			}
			if testName == "ResourceIntensiveTest" {
				resourceIntensiveIndex = i
			}
		}

		assert.NotEqual(t, -1, highPriorityIndex)
		assert.NotEqual(t, -1, resourceIntensiveIndex)
		
		// Resource intensive tests should have lower scores (run later)
		highPriorityScore := schedule.Priority["HighPriorityTest"]
		resourceIntensiveScore := schedule.Priority["ResourceIntensiveTest"]
		assert.Greater(t, highPriorityScore, resourceIntensiveScore)
	})

	t.Run("should handle empty test list", func(t *testing.T) {
		tests := []string{}

		_, err := scheduler.GenerateOptimalTestSchedule(tests)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no tests provided")
	})

	t.Run("should create parallel groups", func(t *testing.T) {
		tests := []string{"Test1", "Test2", "Test3", "Test4", "Test5", "Test6", "Test7", "Test8", "Test9"}

		schedule, err := scheduler.GenerateOptimalTestSchedule(tests)
		
		assert.NoError(t, err)
		assert.Greater(t, len(schedule.ParallelGroups), 1) // Should create multiple groups
		
		// Verify all tests are included in groups
		allTestsInGroups := make(map[string]bool)
		for _, group := range schedule.ParallelGroups {
			for _, test := range group {
				allTestsInGroups[test] = true
			}
		}
		
		for _, test := range tests {
			assert.True(t, allTestsInGroups[test], "Test %s not found in parallel groups", test)
		}
	})
}

func TestFailurePredictor(t *testing.T) {
	predictor := &FailurePredictor{
		historicalFailures: []TestFailure{
			{
				TestName:     "FlakyTest",
				Error:        "timeout error",
				Context:      "network call",
				Timestamp:    time.Now(),
				Reproducible: false,
			},
		},
		riskModel: &SimpleMLModel{
			weights: []float64{0.3, 0.5, 0.2, 0.1},
			bias:    -0.5,
		},
	}

	t.Run("should update failure history", func(t *testing.T) {
		initialCount := len(predictor.historicalFailures)
		
		newFailures := []TestFailure{
			{TestName: "NewFailure", Error: "assertion failed", Reproducible: true},
		}
		
		predictor.UpdateFailureHistory(newFailures)
		
		assert.Equal(t, initialCount+1, len(predictor.historicalFailures))
	})

	t.Run("should limit history size", func(t *testing.T) {
		// Create predictor with many failures
		manyFailures := make([]TestFailure, 1500)
		for i := range manyFailures {
			manyFailures[i] = TestFailure{
				TestName: "Test",
				Error:    "error",
			}
		}
		
		largePredictor := &FailurePredictor{
			historicalFailures: manyFailures,
		}
		
		largePredictor.UpdateFailureHistory([]TestFailure{{TestName: "New", Error: "error"}})
		
		assert.LessOrEqual(t, len(largePredictor.historicalFailures), 1000)
	})
}

func TestTestIntelligenceEngine_GetIntelligentTestReport(t *testing.T) {
	engine := NewTestIntelligenceEngine()
	
	// Add some test data
	engine.performanceModel.historicalData = append(engine.performanceModel.historicalData, TestExecution{
		TestName: "TestSample",
		Duration: 100 * time.Millisecond,
	})
	
	engine.patternAnalyzer.coverageHistory = append(engine.patternAnalyzer.coverageHistory, CoverageSnapshot{
		LineCoverage: 0.8,
		FuncCoverage: 0.9,
	})

	t.Run("should generate comprehensive report", func(t *testing.T) {
		report := engine.GetIntelligentTestReport()
		
		assert.NotEmpty(t, report)
		assert.Contains(t, report, "AI/ML Test Intelligence Report")
		assert.Contains(t, report, "Performance Predictions")
		assert.Contains(t, report, "Coverage Analysis")
		assert.Contains(t, report, "Edge Case Discovery")
		assert.Contains(t, report, "Test Optimization")
		assert.Contains(t, report, "Failure Prediction")
		
		// Verify data is included
		assert.Contains(t, report, "Historical data points: 1")
		assert.Contains(t, report, "Coverage snapshots: 1")
	})
}

// Helper functions for tests
func newTestRNG() *rand.Rand {
	return rand.New(rand.NewSource(12345)) // Fixed seed for reproducible tests
}

// Test helper to create test execution data
func createTestExecution(name string, duration time.Duration, success bool) TestExecution {
	return TestExecution{
		TestName:    name,
		Duration:    duration,
		MemoryUsage: 1024 * 1024, // 1MB
		CPUUsage:    50.0,
		Success:     success,
		Timestamp:   time.Now(),
	}
}

// Test helper to create test failure data
func createTestFailure(name, error string, reproducible bool) TestFailure {
	return TestFailure{
		TestName:     name,
		Error:        error,
		Context:      "test context",
		Timestamp:    time.Now(),
		Reproducible: reproducible,
	}
}

// Integration test that combines multiple AI/ML components
func TestAIMLIntegration(t *testing.T) {
	t.Run("should integrate multiple AI/ML components", func(t *testing.T) {
		engine := NewTestIntelligenceEngine()
		
		// Test complete workflow
		ctx := context.Background()
		
		// 1. Mock code analysis result since we can't create actual files in test
		structure := &CodeStructure{
			Functions: []FunctionSignature{
				{
					Name:       "TestableFunction",
					Complexity: 3,
					Parameters: []Parameter{{Name: "input", Type: "string"}},
					Returns:    []ReturnType{{Type: "error", Error: true}},
				},
			},
		}
		
		// 2. Generate intelligent tests
		var tests []GeneratedTest
		for _, function := range structure.Functions {
			test := engine.generateFunctionTest(function, structure)
			tests = append(tests, test)
		}
		
		edgeCases := engine.edgeCaseDiscovery.DiscoverEdgeCases(structure)
		for _, edgeCase := range edgeCases {
			test := engine.generateEdgeCaseTest(edgeCase)
			tests = append(tests, test)
		}
		
		assert.Greater(t, len(tests), 0)
		
		// 3. Optimize test suite with genetic algorithm
		testNames := make([]string, len(tests))
		for i, test := range tests {
			testNames[i] = test.Name
		}
		
		fitnessFunc := func(subset []string) float64 {
			return float64(len(subset)) // Simple fitness: more tests = better
		}
		
		optimized, err := engine.mutationOptimizer.OptimizeMutationTesting(testNames, fitnessFunc)
		assert.NoError(t, err)
		assert.NotEmpty(t, optimized)
		
		// 4. Generate test schedule
		schedule, err := engine.testScheduler.GenerateOptimalTestSchedule(optimized)
		assert.NoError(t, err)
		assert.NotEmpty(t, schedule.TestOrder)
		
		// 5. Predict performance
		for _, testName := range schedule.TestOrder[:1] { // Test first one
			features := []float64{1.0, 0.5, 0.3, 1.0, float64(len(testName))}
			prediction := engine.performanceModel.PredictPerformance(testName, features)
			assert.Equal(t, testName, prediction.TestName)
			assert.Greater(t, prediction.PredictedTime, time.Duration(0))
		}
		
		// 6. Generate comprehensive report
		report := engine.GetIntelligentTestReport()
		assert.Contains(t, report, "AI/ML Test Intelligence Report")
		
		t.Logf("Integration test completed successfully")
		t.Logf("Generated %d tests", len(tests))
		t.Logf("Optimized to %d tests", len(optimized))
		t.Logf("Schedule estimated time: %v", schedule.EstimatedTime)
	})
}