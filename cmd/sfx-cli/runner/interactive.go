package runner

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/fatih/color"
	"github.com/fsnotify/fsnotify"
	"github.com/manifoldco/promptui"
)

// TestResult represents the result of a test execution
type TestResult struct {
	Package    string
	Name       string
	Status     TestStatus
	Duration   time.Duration
	Error      string
	Output     string
	Coverage   float64
	Timestamp  time.Time
}

// TestStatus represents the status of a test
type TestStatus int

const (
	TestPending TestStatus = iota
	TestRunning
	TestPassed
	TestFailed
	TestSkipped
)

func (s TestStatus) String() string {
	switch s {
	case TestPending:
		return "PENDING"
	case TestRunning:
		return "RUNNING"
	case TestPassed:
		return "PASSED"
	case TestFailed:
		return "FAILED"
	case TestSkipped:
		return "SKIPPED"
	default:
		return "UNKNOWN"
	}
}

// InteractiveRunner provides real-time test execution with feedback
type InteractiveRunner struct {
	projectRoot   string
	filter        string
	watchMode     bool
	coverage      bool
	interactive   bool
	timeout       string
	
	results       map[string]*TestResult
	resultsMutex  sync.RWMutex
	
	watcher       *fsnotify.Watcher
	ctx           context.Context
	cancel        context.CancelFunc
	
	// UI components
	green         *color.Color
	red           *color.Color
	yellow        *color.Color
	blue          *color.Color
	cyan          *color.Color
	bold          *color.Color
}

// NewInteractiveRunner creates a new interactive test runner
func NewInteractiveRunner(projectRoot, filter, timeout string, watchMode, coverage, interactive bool) (*InteractiveRunner, error) {
	ctx, cancel := context.WithCancel(context.Background())
	
	var watcher *fsnotify.Watcher
	var err error
	
	if watchMode {
		watcher, err = fsnotify.NewWatcher()
		if err != nil {
			cancel()
			return nil, fmt.Errorf("failed to create file watcher: %w", err)
		}
	}
	
	return &InteractiveRunner{
		projectRoot: projectRoot,
		filter:      filter,
		watchMode:   watchMode,
		coverage:    coverage,
		interactive: interactive,
		timeout:     timeout,
		results:     make(map[string]*TestResult),
		watcher:     watcher,
		ctx:         ctx,
		cancel:      cancel,
		
		// Initialize colors
		green:  color.New(color.FgGreen),
		red:    color.New(color.FgRed),
		yellow: color.New(color.FgYellow),
		blue:   color.New(color.FgBlue),
		cyan:   color.New(color.FgCyan),
		bold:   color.New(color.Bold),
	}, nil
}

// Run starts the interactive test runner
func (r *InteractiveRunner) Run() error {
	defer r.cleanup()
	
	r.printBanner()
	
	if r.interactive {
		return r.runInteractiveMode()
	}
	
	if r.watchMode {
		return r.runWatchMode()
	}
	
	return r.runOnce()
}

// printBanner displays the SFX test runner banner
func (r *InteractiveRunner) printBanner() {
	r.cyan.Println(`
╔═══════════════════════════════════════════════════════════════════════════════╗
║                           SFX Interactive Test Runner                        ║
║                        Serverless Framework eXtended                         ║
╚═══════════════════════════════════════════════════════════════════════════════╝
`)
	
	fmt.Printf("📍 Project: %s\n", r.projectRoot)
	if r.filter != "" {
		fmt.Printf("🔍 Filter: %s\n", r.filter)
	}
	if r.watchMode {
		r.yellow.Println("👀 Watch mode enabled - tests will rerun on file changes")
	}
	if r.coverage {
		fmt.Println("📊 Coverage analysis enabled")
	}
	fmt.Println()
}

// runInteractiveMode runs the interactive mode with menu selection
func (r *InteractiveRunner) runInteractiveMode() error {
	for {
		action, err := r.showMainMenu()
		if err != nil {
			return err
		}
		
		switch action {
		case "run_all":
			if err := r.runAllTests(); err != nil {
				r.red.Printf("Error running tests: %v\n", err)
			}
		case "run_package":
			if err := r.runPackageTests(); err != nil {
				r.red.Printf("Error running package tests: %v\n", err)
			}
		case "run_specific":
			if err := r.runSpecificTest(); err != nil {
				r.red.Printf("Error running specific test: %v\n", err)
			}
		case "watch":
			return r.runWatchMode()
		case "coverage":
			if err := r.showCoverageReport(); err != nil {
				r.red.Printf("Error showing coverage: %v\n", err)
			}
		case "results":
			r.showTestResults()
		case "clear":
			r.clearResults()
		case "exit":
			return nil
		}
		
		fmt.Println()
	}
}

// showMainMenu displays the interactive menu
func (r *InteractiveRunner) showMainMenu() (string, error) {
	prompt := promptui.Select{
		Label: "Select action",
		Items: []string{
			"🧪 Run All Tests",
			"📦 Run Package Tests",
			"🎯 Run Specific Test",
			"👀 Enable Watch Mode", 
			"📊 Show Coverage Report",
			"📋 Show Test Results",
			"🧹 Clear Results",
			"❌ Exit",
		},
		Templates: &promptui.SelectTemplates{
			Label:    "{{ . }}",
			Active:   "▶ {{ . | cyan }}",
			Inactive: "  {{ . }}",
			Selected: "✓ {{ . | green }}",
		},
	}
	
	index, _, err := prompt.Run()
	if err != nil {
		return "", err
	}
	
	actions := []string{"run_all", "run_package", "run_specific", "watch", "coverage", "results", "clear", "exit"}
	return actions[index], nil
}

// runWatchMode runs tests in watch mode with file system monitoring
func (r *InteractiveRunner) runWatchMode() error {
	r.cyan.Println("🔄 Starting watch mode...")
	
	// Setup file watching
	if err := r.setupFileWatching(); err != nil {
		return fmt.Errorf("failed to setup file watching: %w", err)
	}
	
	// Run tests initially
	if err := r.runAllTests(); err != nil {
		r.red.Printf("Initial test run failed: %v\n", err)
	}
	
	// Start watching for changes
	for {
		select {
		case event, ok := <-r.watcher.Events:
			if !ok {
				return nil
			}
			
			if r.shouldRunTests(event) {
				r.cyan.Printf("📁 File changed: %s\n", event.Name)
				r.yellow.Println("🔄 Rerunning tests...")
				
				if err := r.runAllTests(); err != nil {
					r.red.Printf("Test run failed: %v\n", err)
				}
			}
			
		case err, ok := <-r.watcher.Errors:
			if !ok {
				return nil
			}
			r.red.Printf("Watch error: %v\n", err)
			
		case <-r.ctx.Done():
			return nil
		}
	}
}

// runOnce runs tests once and exits
func (r *InteractiveRunner) runOnce() error {
	return r.runAllTests()
}

// runAllTests executes all tests in the project
func (r *InteractiveRunner) runAllTests() error {
	r.cyan.Println("🧪 Running all tests...")
	
	packages, err := r.discoverTestPackages()
	if err != nil {
		return fmt.Errorf("failed to discover test packages: %w", err)
	}
	
	if len(packages) == 0 {
		r.yellow.Println("⚠️  No test packages found")
		return nil
	}
	
	r.blue.Printf("📦 Found %d test packages\n", len(packages))
	
	var wg sync.WaitGroup
	for _, pkg := range packages {
		wg.Add(1)
		go func(packageName string) {
			defer wg.Done()
			r.runPackageTest(packageName)
		}(pkg)
	}
	
	wg.Wait()
	r.showSummary()
	return nil
}

// runPackageTests allows user to select and run specific package tests
func (r *InteractiveRunner) runPackageTests() error {
	packages, err := r.discoverTestPackages()
	if err != nil {
		return fmt.Errorf("failed to discover test packages: %w", err)
	}
	
	if len(packages) == 0 {
		r.yellow.Println("⚠️  No test packages found")
		return nil
	}
	
	prompt := promptui.Select{
		Label: "Select package to test",
		Items: packages,
	}
	
	_, selectedPackage, err := prompt.Run()
	if err != nil {
		return err
	}
	
	r.cyan.Printf("🧪 Running tests for package: %s\n", selectedPackage)
	return r.runPackageTest(selectedPackage)
}

// runSpecificTest allows user to run a specific test
func (r *InteractiveRunner) runSpecificTest() error {
	prompt := promptui.Prompt{
		Label: "Enter test name pattern (e.g., TestFunctionName)",
	}
	
	testPattern, err := prompt.Run()
	if err != nil {
		return err
	}
	
	r.cyan.Printf("🎯 Running tests matching: %s\n", testPattern)
	return r.runTestWithPattern(testPattern)
}

// runPackageTest executes tests for a specific package
func (r *InteractiveRunner) runPackageTest(packageName string) error {
	startTime := time.Now()
	
	// Update result status to running
	r.updateTestResult(packageName, &TestResult{
		Package:   packageName,
		Status:    TestRunning,
		Timestamp: startTime,
	})
	
	// Build command
	args := []string{"test", "-v"}
	
	if r.coverage {
		coverageFile := filepath.Join("coverage", fmt.Sprintf("%s.cov", packageName))
		args = append(args, "-coverprofile="+coverageFile)
	}
	
	if r.timeout != "" {
		args = append(args, "-timeout="+r.timeout)
	}
	
	if r.filter != "" {
		args = append(args, "-run="+r.filter)
	}
	
	args = append(args, "./"+packageName)
	
	cmd := exec.CommandContext(r.ctx, "go", args...)
	cmd.Dir = r.projectRoot
	
	output, err := cmd.CombinedOutput()
	duration := time.Since(startTime)
	
	result := &TestResult{
		Package:   packageName,
		Duration:  duration,
		Output:    string(output),
		Timestamp: time.Now(),
	}
	
	if err != nil {
		result.Status = TestFailed
		result.Error = err.Error()
		r.red.Printf("❌ %s failed (%v)\n", packageName, duration.Truncate(time.Millisecond))
	} else {
		result.Status = TestPassed
		r.green.Printf("✅ %s passed (%v)\n", packageName, duration.Truncate(time.Millisecond))
	}
	
	// Extract coverage if available
	if r.coverage {
		coverage, _ := r.extractCoverage(packageName)
		result.Coverage = coverage
	}
	
	r.updateTestResult(packageName, result)
	return nil
}

// runTestWithPattern runs tests matching a specific pattern
func (r *InteractiveRunner) runTestWithPattern(pattern string) error {
	cmd := exec.CommandContext(r.ctx, "go", "test", "-v", "-run="+pattern, "./...")
	cmd.Dir = r.projectRoot
	
	output, err := cmd.CombinedOutput()
	
	if err != nil {
		r.red.Printf("❌ Tests matching '%s' failed\n", pattern)
		fmt.Println(string(output))
	} else {
		r.green.Printf("✅ Tests matching '%s' passed\n", pattern)
	}
	
	return nil
}

// discoverTestPackages finds all packages with test files
func (r *InteractiveRunner) discoverTestPackages() ([]string, error) {
	var packages []string
	
	err := filepath.Walk(r.projectRoot, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip errors
		}
		
		if !info.IsDir() && strings.HasSuffix(info.Name(), "_test.go") {
			dir := filepath.Dir(path)
			relDir, err := filepath.Rel(r.projectRoot, dir)
			if err != nil {
				return nil
			}
			
			// Skip vendor and hidden directories
			if strings.Contains(relDir, "vendor/") || strings.HasPrefix(relDir, ".") {
				return nil
			}
			
			// Add package if not already added
			found := false
			for _, pkg := range packages {
				if pkg == relDir {
					found = true
					break
				}
			}
			if !found {
				packages = append(packages, relDir)
			}
		}
		
		return nil
	})
	
	return packages, err
}

// setupFileWatching sets up file system monitoring for watch mode
func (r *InteractiveRunner) setupFileWatching() error {
	return filepath.Walk(r.projectRoot, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		
		if info.IsDir() && !strings.Contains(path, "/.") && !strings.Contains(path, "/vendor/") {
			return r.watcher.Add(path)
		}
		
		return nil
	})
}

// shouldRunTests determines if tests should run based on file change
func (r *InteractiveRunner) shouldRunTests(event fsnotify.Event) bool {
	if event.Op&fsnotify.Write == fsnotify.Write {
		ext := filepath.Ext(event.Name)
		return ext == ".go" || ext == ".mod" || ext == ".sum"
	}
	return false
}

// updateTestResult updates the test result in thread-safe manner
func (r *InteractiveRunner) updateTestResult(packageName string, result *TestResult) {
	r.resultsMutex.Lock()
	defer r.resultsMutex.Unlock()
	r.results[packageName] = result
}

// showTestResults displays all test results
func (r *InteractiveRunner) showTestResults() {
	r.resultsMutex.RLock()
	defer r.resultsMutex.RUnlock()
	
	if len(r.results) == 0 {
		r.yellow.Println("⚠️  No test results available")
		return
	}
	
	r.bold.Println("📋 Test Results:")
	fmt.Println(strings.Repeat("─", 80))
	
	for _, result := range r.results {
		status := ""
		switch result.Status {
		case TestPassed:
			status = r.green.Sprint("✅ PASSED")
		case TestFailed:
			status = r.red.Sprint("❌ FAILED")
		case TestRunning:
			status = r.yellow.Sprint("🔄 RUNNING")
		case TestSkipped:
			status = r.cyan.Sprint("⏭️  SKIPPED")
		default:
			status = "❓ PENDING"
		}
		
		fmt.Printf("%-30s %s (%v)", result.Package, status, result.Duration.Truncate(time.Millisecond))
		
		if r.coverage && result.Coverage > 0 {
			fmt.Printf(" [%.1f%% coverage]", result.Coverage)
		}
		
		fmt.Println()
		
		if result.Error != "" {
			r.red.Printf("   Error: %s\n", result.Error)
		}
	}
	
	fmt.Println(strings.Repeat("─", 80))
}

// showSummary displays a summary of test results
func (r *InteractiveRunner) showSummary() {
	r.resultsMutex.RLock()
	defer r.resultsMutex.RUnlock()
	
	var passed, failed, total int
	var totalDuration time.Duration
	
	for _, result := range r.results {
		total++
		totalDuration += result.Duration
		
		switch result.Status {
		case TestPassed:
			passed++
		case TestFailed:
			failed++
		}
	}
	
	fmt.Println()
	r.bold.Println("📊 Test Summary:")
	fmt.Println(strings.Repeat("═", 50))
	
	if passed > 0 {
		r.green.Printf("✅ Passed: %d\n", passed)
	}
	if failed > 0 {
		r.red.Printf("❌ Failed: %d\n", failed)
	}
	
	fmt.Printf("📦 Total packages: %d\n", total)
	fmt.Printf("⏱️  Total duration: %v\n", totalDuration.Truncate(time.Millisecond))
	
	if failed == 0 {
		r.green.Println("🎉 All tests passed!")
	} else {
		r.red.Println("💥 Some tests failed!")
	}
	
	fmt.Println(strings.Repeat("═", 50))
}

// clearResults clears all test results
func (r *InteractiveRunner) clearResults() {
	r.resultsMutex.Lock()
	defer r.resultsMutex.Unlock()
	
	r.results = make(map[string]*TestResult)
	r.cyan.Println("🧹 Results cleared")
}

// showCoverageReport displays coverage information
func (r *InteractiveRunner) showCoverageReport() error {
	if !r.coverage {
		r.yellow.Println("⚠️  Coverage not enabled")
		return nil
	}
	
	r.bold.Println("📊 Coverage Report:")
	fmt.Println(strings.Repeat("─", 60))
	
	r.resultsMutex.RLock()
	defer r.resultsMutex.RUnlock()
	
	for packageName, result := range r.results {
		if result.Coverage > 0 {
			coverage := result.Coverage
			status := ""
			
			if coverage >= 90 {
				status = r.green.Sprint("🟢")
			} else if coverage >= 70 {
				status = r.yellow.Sprint("🟡")
			} else {
				status = r.red.Sprint("🔴")
			}
			
			fmt.Printf("%s %-30s %6.1f%%\n", status, packageName, coverage)
		}
	}
	
	fmt.Println(strings.Repeat("─", 60))
	return nil
}

// extractCoverage extracts coverage percentage from coverage files
func (r *InteractiveRunner) extractCoverage(packageName string) (float64, error) {
	coverageFile := filepath.Join(r.projectRoot, "coverage", fmt.Sprintf("%s.cov", packageName))
	
	if _, err := os.Stat(coverageFile); os.IsNotExist(err) {
		return 0, nil
	}
	
	cmd := exec.Command("go", "tool", "cover", "-func="+coverageFile)
	output, err := cmd.Output()
	if err != nil {
		return 0, err
	}
	
	// Parse the last line which contains total coverage
	lines := strings.Split(string(output), "\n")
	for i := len(lines) - 1; i >= 0; i-- {
		if strings.Contains(lines[i], "total:") {
			parts := strings.Fields(lines[i])
			if len(parts) >= 3 {
				coverageStr := strings.TrimSuffix(parts[2], "%")
				var coverage float64
				fmt.Sscanf(coverageStr, "%f", &coverage)
				return coverage, nil
			}
		}
	}
	
	return 0, nil
}

// cleanup performs cleanup when runner is stopped
func (r *InteractiveRunner) cleanup() {
	if r.watcher != nil {
		r.watcher.Close()
	}
	r.cancel()
}

// WaitForUserInput waits for user input to continue
func (r *InteractiveRunner) WaitForUserInput() {
	fmt.Print("Press Enter to continue...")
	bufio.NewReader(os.Stdin).ReadBytes('\n')
}